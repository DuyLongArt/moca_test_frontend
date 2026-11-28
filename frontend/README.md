# MoCA Platform — Frontend

React 19 + Vite + TypeScript. Clinical minimal UI.

## Quick start

```bash
npm install
npm run dev
```

Open http://localhost:5173 — chọn vai trò trên màn đăng nhập demo.

## Stack

| Layer | Path |
|-------|------|
| Routes | `src/router.tsx` |
| Features | `src/features/{patient,doctor,admin,auth}/` |
| Link (HTTP) | `src/shared/lib/axios.ts` |
| Client state | `src/stores/authStore.ts` (Zustand) |
| Server state | TanStack Query (`src/features/*/use*Queries.ts`) |
| MoCA wizard | `src/features/patient/MocaTestContainer.jsx` |

## API proxy

Vite proxies `/api` → the deployed Worker (`https://mocatestdeploy.duylongart.workers.dev` by default). Set `VITE_PROXY_TARGET=http://localhost:8787` to use a local `wrangler dev` server.

Copy `.env.example` → `.env` if needed.

## Design

See [`../design.md`](../design.md) before UI changes.
