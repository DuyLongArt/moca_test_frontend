import { DurableObject } from "cloudflare:workers";
import { Hono } from "hono";

const PORT = 8080;
const INACTIVITY_TIMEOUT_MS = 10 * 60 * 1000;
const READY_PATH = "/api/v1/health";
const INSTANCE_NAME = "moca";
const FORGOT_PASSWORD_PATH = "/api/v1/auth/forgot-password";
const INTERNAL_FORGOT_PATH = "/internal/password-reset/request";
const INTERNAL_D1_PATH = "/internal/d1/query";
const GENERIC_FORGOT_MESSAGE =
	"If an account exists for that email, a password reset link has been sent.";

function parseEmail(value: unknown): string | null {
	if (typeof value !== "string") return null;
	const email = value.trim().toLowerCase();
	return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email) ? email : null;
}

function escapeHtml(value: string): string {
	return value.replace(
		/[&<>"']/g,
		(character) =>
			({
				"&": "&amp;",
				"<": "&lt;",
				">": "&gt;",
				'"': "&quot;",
				"'": "&#39;",
			})[character] ?? character,
	);
}

/** Env vars forwarded into the MoCA container (set via wrangler vars/secrets). */
function containerEnv(env: Env, instanceId: string): Record<string, string> {
	const keys = [
		"SPRING_PROFILES_ACTIVE",
		"DB_HOST",
		"DB_PORT",
		"DB_NAME",
		"DB_USER",
		"DB_PASSWORD",
		"JWT_SECRET",
		"D1_PROXY_URL",
		"D1_PROXY_TOKEN",
		"REDIS_ENABLED",
		"REDIS_HOST",
		"REDIS_PORT",
		"REDIS_PASSWORD",
		"REDIS_SSL",
		"MINIO_ENABLED",
		"MINIO_ENDPOINT",
		"MINIO_ACCESS_KEY",
		"MINIO_SECRET_KEY",
		"MINIO_BUCKET",
		"WHISPER_ENABLED",
		"WHISPER_PROVIDER",
		"WHISPER_BASE_URL",
		"WHISPER_API_KEY",
		"WHISPER_MODEL",
		"WHISPER_LANGUAGE",
	] as const;

	const out: Record<string, string> = {
		INSTANCE_ID: instanceId,
		SPRING_PROFILES_ACTIVE: "prod",
		REDIS_ENABLED: "false",
		MINIO_ENABLED: "false",
		WHISPER_ENABLED: "false",
	};

	const bag = env as Env & Record<string, string | undefined>;
	for (const key of keys) {
		const value = bag[key];
		if (typeof value === "string" && value.length > 0) {
			out[key] = value;
		}
	}
	return out;
}

export class MyContainer extends DurableObject<Env> {
	private starting: Promise<void> | undefined;
	private monitoring: Promise<void> | undefined;

	constructor(ctx: DurableObjectState, env: Env) {
		super(ctx, env);
		const container = ctx.container;
		if (container?.running) {
			void ctx.blockConcurrencyWhile(() =>
				container.setInactivityTimeout(INACTIVITY_TIMEOUT_MS),
			);
			this.observeExit();
		}
	}

	async fetch(request: Request): Promise<Response> {
		this.starting ??= this.startAndWaitForPort().finally(() => {
			this.starting = undefined;
		});
		await this.starting;

		const url = new URL(request.url);
		url.protocol = "http:";
		url.host = "container";
		const forwarded = new Request(url, request);
		forwarded.headers.delete("host");
		return this.ctx.container!.getTcpPort(PORT).fetch(forwarded);
	}

	private async startAndWaitForPort(): Promise<void> {
		const container = this.ctx.container!;
		if (!container.running) {
			container.start({
				image: container.images.base,
				// JVM + nginx need more than lite
				instance: "standard-2",
				enableInternet: true,
				env: containerEnv(this.env, this.ctx.id.toString()),
			});
		}
		this.observeExit();
		await container.setInactivityTimeout(INACTIVITY_TIMEOUT_MS);

		const port = container.getTcpPort(PORT);
		const deadline = Date.now() + 30_000;
		while (Date.now() < deadline) {
			try {
				const response = await port.fetch(`http://container${READY_PATH}`);
				await response.body?.cancel();
				if (response.ok) return;
			} catch {}
			await new Promise((resolve) => setTimeout(resolve, 200));
		}
		throw new Error("MoCA container did not become ready on port 8080");
	}

	private observeExit(): void {
		if (this.monitoring) return;
		this.monitoring = this.ctx
			.container!.monitor()
			.then(() => console.log("MoCA container exited successfully"))
			.catch((error: unknown) => console.error("MoCA container failed:", error))
			.finally(() => {
				this.monitoring = undefined;
			});
		this.ctx.waitUntil(this.monitoring);
	}
}

const app = new Hono<{ Bindings: Env }>();

type D1StatementInput = { sql: string; params?: unknown[] };

app.post(INTERNAL_D1_PATH, async (c) => {
	const env = c.env as Env & { D1_PROXY_TOKEN?: string };
	if (!env.D1_PROXY_TOKEN || c.req.header("x-d1-proxy-token") !== env.D1_PROXY_TOKEN) {
		return c.json({ success: false, error: "Unauthorized" }, 401);
	}

	let body: unknown;
	try {
		body = await c.req.json();
	} catch {
		return c.json({ success: false, error: "Cannot parse JSON payload" }, 400);
	}
	if (!body || typeof body !== "object") {
		return c.json({ success: false, error: "Invalid D1 request" }, 400);
	}

	const payload = body as {
		sql?: unknown;
		params?: unknown;
		raw?: unknown;
		batch?: unknown;
	};
	const db = c.env.moca_db_lite;
	const prepare = (input: unknown): D1PreparedStatement => {
		if (!input || typeof input !== "object") throw new Error("Invalid statement");
		const { sql, params } = input as D1StatementInput;
		if (typeof sql !== "string") throw new Error("SQL is required");
		const values = Array.isArray(params) ? params : [];
		return values.length > 0 ? db.prepare(sql).bind(...values) : db.prepare(sql);
	};

	try {
		if (Array.isArray(payload.batch)) {
			const statements = payload.batch.map(prepare);
			const results = statements.length > 0 ? await db.batch(statements) : [];
			return c.json({
				success: true,
				results: results.map((result) => ({
					success: result.success,
					meta: result.meta,
				})),
			});
		}

		const statement = prepare(body);
		if (payload.raw === true) {
			const rows = await statement.raw<unknown[]>({ columnNames: true });
			return c.json({
				success: true,
				columns: rows[0] ?? [],
				rows: rows.slice(1),
			});
		}

		const result = await statement.run();
		return c.json({ success: result.success, meta: result.meta });
	} catch (error) {
		return c.json(
			{
				success: false,
				error: error instanceof Error ? error.message : String(error),
			},
			400,
		);
	}
});

app.post(FORGOT_PASSWORD_PATH, async (c) => {
	let body: unknown;
	try {
		body = await c.req.json();
	} catch {
		return c.json(
			{ status: "ERROR", code: 400, message: "Cannot parse JSON payload" },
			400,
		);
	}

	const email = parseEmail(
		body && typeof body === "object" && "email" in body
			? (body as { email?: unknown }).email
			: null,
	);
	const role =
		body && typeof body === "object" && "role" in body && (body as { role?: unknown }).role === "patient"
			? "patient"
			: "doctor";
	if (!email) {
		return c.json(
			{ status: "ERROR", code: 400, message: "Invalid email address" },
			400,
		);
	}

	const upstream = await c.env.MY_CONTAINER.getByName(INSTANCE_NAME).fetch(
		new Request(new URL(INTERNAL_FORGOT_PATH, c.req.url), {
			method: "POST",
			headers: { "content-type": "application/json" },
			body: JSON.stringify({ email, role }),
		}),
	);
	if (!upstream.ok) {
		console.error(
			JSON.stringify({
				message: "password reset token request failed",
				status: upstream.status,
				path: INTERNAL_FORGOT_PATH,
			}),
		);
		return c.json(
			{ status: "ERROR", code: 502, message: "Could not start password reset" },
			502,
		);
	}

	let token: string | undefined;
	let resetRole: "doctor" | "patient" = role;
	try {
		const payload = await upstream.json<{ data?: { token?: string; role?: string } }>();
		token = payload.data?.token;
		if (payload.data?.role === "patient" || payload.data?.role === "doctor") {
			resetRole = payload.data.role;
		}
	} catch {
		token = undefined;
	}

	if (token) {
		const resetUrl = `${new URL(c.req.url).origin}/reset-password?token=${encodeURIComponent(token)}&role=${resetRole}`;
		try {
			await c.env.EMAIL.send({
				to: email,
				from: c.env.EMAIL_FROM,
				subject: "Reset your MoCA password",
				text: `Use this link to reset your MoCA password: ${resetUrl}\nThis link expires in 30 minutes.`,
				html: `<p>Use this link to reset your MoCA password:</p><p><a href="${escapeHtml(resetUrl)}">Reset password</a></p><p>This link expires in 30 minutes.</p>`,
			});
		} catch (error) {
			console.error(
				JSON.stringify({
					message: "password reset email failed",
					error: error instanceof Error ? error.message : String(error),
				}),
			);
		}
	}

	return c.json({
		status: "OK",
		code: 200,
		message: GENERIC_FORGOT_MESSAGE,
		data: null,
	});
});

app.all("*", (c) => {
	if (new URL(c.req.url).pathname.startsWith("/internal/")) {
		return c.notFound();
	}
	return c.env.MY_CONTAINER.getByName(INSTANCE_NAME).fetch(c.req.raw);
});

app.onError((error, c) => {
	console.error("MoCA container request failed:", error);
	return c.text(
		"MoCA container request failed. Check Worker logs and DB secrets, then retry.",
		502,
	);
});

export default app;
