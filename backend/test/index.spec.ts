import { afterEach, describe, expect, it, vi } from "vitest";
import worker, { MyContainer } from "../src/index";

function createEnv() {
	const fetch = vi.fn(async () => new Response("container response"));
	const getByName = vi.fn(() => ({ fetch }));
	const sendEmail = vi.fn(async () => ({ messageId: "test-message-id" }));
	const env = {
		MY_CONTAINER: { getByName },
		EMAIL: { send: sendEmail },
		EMAIL_FROM: "no-reply@example.com",
		DB_HOST: "db.example",
		DB_PORT: "5432",
		DB_NAME: "moca-test-db",
		DB_USER: "duylong",
		DB_PASSWORD: "secret",
		JWT_SECRET: "x".repeat(48),
	} as unknown as Env;
	return { env, fetch, getByName, sendEmail };
}

function createContainer(running = false) {
	const health = vi.fn(async () => new Response(null, { status: 200 }));
	const forward = vi.fn(async (_request: Request) => new Response("upstream"));
	let resolveExit: () => void;
	let rejectExit: (error: Error) => void;
	const container = {
		running,
		images: { base: "registry.cloudflare.com/example/image@sha256:abc" },
		start: vi.fn(() => {
			container.running = true;
		}),
		setInactivityTimeout: vi.fn(async () => {
			expect(container.running).toBe(true);
		}),
		monitor: vi.fn(
			() =>
				new Promise<void>((resolve, reject) => {
					resolveExit = resolve;
					rejectExit = reject;
				}),
		),
		getTcpPort: vi.fn(() => ({
			fetch: (request: Request | string) =>
				typeof request === "string" ? health() : forward(request),
		})),
	};
	const waitUntil = vi.fn();
	const object: MyContainer = Object.create(MyContainer.prototype);
	Object.defineProperty(object, "ctx", {
		value: { container, waitUntil, id: { toString: () => "test-instance" } },
	});
	Object.defineProperty(object, "env", {
		value: {
			DB_HOST: "db.example",
			JWT_SECRET: "x".repeat(48),
		},
	});
	return {
		object,
		container,
		health,
		forward,
		waitUntil,
		exit: () => {
			container.running = false;
			resolveExit();
		},
		fail: () => {
			container.running = false;
			rejectExit(new Error("exit code 1"));
		},
	};
}

function createD1Env() {
	const statement = {
		bind: vi.fn(() => statement),
		run: vi.fn(async () => ({ success: true, meta: { changes: 1, last_row_id: 7 } })),
		raw: vi.fn(async () => [["id", "email"], [7, "doctor@example.com"]]),
	};
	const db = {
		prepare: vi.fn(() => statement),
		batch: vi.fn(async (statements: unknown[]) =>
			statements.map(() => ({ success: true, meta: { changes: 1 } })),
		),
	};
	const env = {
		...createEnv().env,
		moca_db_lite: db,
		D1_PROXY_TOKEN: "d1-secret",
	} as unknown as Env;
	return { env, db, statement };
}

afterEach(() => vi.restoreAllMocks());

describe("Worker routes", () => {
	it("rejects unauthenticated D1 proxy requests", async () => {
		const { env } = createD1Env();
		const response = await worker.request("http://example.com/internal/d1/query", {
			method: "POST",
			headers: { "content-type": "application/json" },
			body: JSON.stringify({ sql: "SELECT 1" }),
		}, env);
		expect(response.status).toBe(401);
	});

	it("executes D1 queries through the binding", async () => {
		const { env, db, statement } = createD1Env();
		const response = await worker.request("http://example.com/internal/d1/query", {
			method: "POST",
			headers: {
				"content-type": "application/json",
				"x-d1-proxy-token": "d1-secret",
			},
			body: JSON.stringify({ sql: "SELECT id, email FROM doctors WHERE id = ?", params: [7], raw: true }),
		}, env);
		expect(response.status).toBe(200);
		await expect(response.json()).resolves.toMatchObject({
			success: true,
			columns: ["id", "email"],
		});
		expect(db.prepare).toHaveBeenCalledWith("SELECT id, email FROM doctors WHERE id = ?");
		expect(statement.bind).toHaveBeenCalledWith(7);
		expect(statement.raw).toHaveBeenCalledWith({ columnNames: true });
	});

	it("sends all paths to the shared moca container", async () => {
		const { env, getByName } = createEnv();
		await worker.request("http://example.com/", undefined, env);
		await worker.request("http://example.com/api/v1/health", undefined, env);
		expect(getByName.mock.calls).toEqual([["moca"], ["moca"]]);
	});

	it("returns a useful 502 when the container proxy fails", async () => {
		vi.spyOn(console, "error").mockImplementation(() => {});
		const { env, fetch } = createEnv();
		fetch.mockRejectedValue(new Error("container exited"));
		const response = await worker.request("http://example.com/", undefined, env);
		expect(response.status).toBe(502);
		expect(await response.text()).toContain("MoCA container request failed");
	});

	it("sends a reset email without exposing the token in the response", async () => {
		const { env, fetch, sendEmail } = createEnv();
		fetch.mockResolvedValue(
			new Response(
				JSON.stringify({ status: "OK", data: { token: "reset-token" } }),
				{ headers: { "content-type": "application/json" } },
			),
		);
		const response = await worker.request(
			"http://example.com/api/v1/auth/forgot-password",
			{
				method: "POST",
				headers: { "content-type": "application/json" },
				body: JSON.stringify({ email: " Doctor@Example.com " }),
			},
			env,
		);
		expect(response.status).toBe(200);
		const body: unknown = await response.json();
		expect(body).toMatchObject({ status: "OK", code: 200 });
		expect(JSON.stringify(body)).not.toContain("reset-token");

		const internalRequest = fetch.mock.calls[0][0] as Request;
		expect(new URL(internalRequest.url).pathname).toBe(
			"/internal/password-reset/request",
		);
		expect(await internalRequest.json()).toEqual({
			email: "doctor@example.com",
			role: "doctor",
		});

		expect(sendEmail).toHaveBeenCalledWith(
			expect.objectContaining({
				to: "doctor@example.com",
				from: "no-reply@example.com",
			}),
		);
		const message = sendEmail.mock.calls[0][0] as { text: string };
		expect(message.text).toContain(
			"http://example.com/reset-password?token=reset-token",
		);
	});

	it("keeps the patient role in the reset link", async () => {
		const { env, fetch, sendEmail } = createEnv();
		fetch.mockResolvedValue(
			new Response(
				JSON.stringify({
					status: "OK",
					data: { token: "patient-token", role: "patient" },
				}),
				{ headers: { "content-type": "application/json" } },
			),
		);
		const response = await worker.request(
			"http://example.com/api/v1/auth/forgot-password",
			{
				method: "POST",
				headers: { "content-type": "application/json" },
				body: JSON.stringify({ email: "patient@example.com", role: "patient" }),
			},
			env,
		);
		expect(response.status).toBe(200);

		const internalRequest = fetch.mock.calls[0][0] as Request;
		expect(await internalRequest.json()).toEqual({
			email: "patient@example.com",
			role: "patient",
		});
		const message = sendEmail.mock.calls[0][0] as { text: string };
		expect(message.text).toContain("role=patient");
	});

	it("does not send email when the doctor is not found", async () => {
		const { env, fetch, sendEmail } = createEnv();
		fetch.mockResolvedValue(
			new Response(JSON.stringify({ status: "OK" }), {
				headers: { "content-type": "application/json" },
			}),
		);
		const response = await worker.request(
			"http://example.com/api/v1/auth/forgot-password",
			{
				method: "POST",
				headers: { "content-type": "application/json" },
				body: JSON.stringify({ email: "missing@example.com" }),
			},
			env,
		);
		expect(response.status).toBe(200);
		expect(sendEmail).not.toHaveBeenCalled();
	});

	it("does not expose internal routes through the Worker", async () => {
		const { env, getByName } = createEnv();
		const response = await worker.request(
			"http://example.com/internal/password-reset/request",
			undefined,
			env,
		);
		expect(response.status).toBe(404);
		expect(getByName).not.toHaveBeenCalled();
	});

	it("does not forward the container error route", async () => {
		const { env, getByName } = createEnv();
		const response = await worker.request("http://example.com/error", undefined, env);
		expect(response.status).toBe(404);
		expect(getByName).not.toHaveBeenCalled();
	});
});

describe("Container lifecycle and proxy", () => {
	it("starts MoCA image with internet and standard-2", async () => {
		const { object, container } = createContainer();
		await object.fetch(new Request("https://example.com/"));
		expect(container.start).toHaveBeenCalledWith({
			image: container.images.base,
			instance: "standard-2",
			enableInternet: true,
			env: expect.objectContaining({
				INSTANCE_ID: "test-instance",
				DB_HOST: "db.example",
				JWT_SECRET: "x".repeat(48),
				SPRING_PROFILES_ACTIVE: "prod",
			}),
		});
		expect(container.setInactivityTimeout).toHaveBeenCalledWith(600_000);
	});

	it("waits for /api/v1/health before forwarding", async () => {
		const { object, container, health, forward } = createContainer();
		let ready: (response: Response) => void;
		health.mockImplementation(
			() =>
				new Promise((resolve) => {
					ready = resolve;
				}),
		);
		const first = object.fetch(new Request("https://example.com/one"));
		const second = object.fetch(new Request("https://example.com/two"));
		await vi.waitFor(() => expect(health).toHaveBeenCalledOnce());
		expect(container.start).toHaveBeenCalledOnce();
		expect(forward).not.toHaveBeenCalled();
		ready!(new Response(null, { status: 200 }));
		await Promise.all([first, second]);
		expect(forward).toHaveBeenCalledTimes(2);
	});

	it("preserves method, path, query, body and response", async () => {
		const { object, forward } = createContainer(true);
		const upstream = new Response("result", {
			status: 201,
			headers: { "X-Test": "yes" },
		});
		forward.mockResolvedValue(upstream);
		const response = await object.fetch(
			new Request("https://example.com/data?q=1", {
				method: "POST",
				body: "payload",
				headers: { Host: "example.com" },
			}),
		);
		const request = forward.mock.calls[0][0];
		expect(request.url).toBe("http://container/data?q=1");
		expect(request.method).toBe("POST");
		expect(request.headers.has("host")).toBe(false);
		expect(await request.text()).toBe("payload");
		expect(response).toBe(upstream);
	});

	it("observes failure and restarts on the next request", async () => {
		const logged = vi.spyOn(console, "error").mockImplementation(() => {});
		const { object, container, waitUntil, fail } = createContainer();
		await object.fetch(new Request("https://example.com/one"));
		fail();
		await waitUntil.mock.calls[0][0];
		expect(logged).toHaveBeenCalledWith(
			"MoCA container failed:",
			expect.any(Error),
		);
		await object.fetch(new Request("https://example.com/one"));
		expect(container.start).toHaveBeenCalledTimes(2);
	});
});
