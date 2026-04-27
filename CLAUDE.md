# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is XRPay

XRPay is a payments-as-a-service platform that lets online merchants accept XRP payments. It consists of a Go backend, a React frontend dashboard, and an embeddable JS checkout widget served by the backend.

## Commands

### Backend (Go)
```bash
cd backend
go mod tidy          # install/sync dependencies (one-time)
go run ./main.go     # starts API server on :8080
go build ./...       # compile check without running
go vet ./...         # static analysis
go test ./...        # run all tests
```

### Frontend (React/Vite)
```bash
cd frontend
npm install          # install dependencies (one-time)
npm run dev          # dev server at localhost:5173
npm run build        # production build
npm run lint         # ESLint
```

Both servers must run simultaneously for full functionality. The frontend proxies nothing — it calls `http://localhost:8080` directly.

## Architecture

### Request flow
```
Browser/Widget → Go HTTP server (port 8080)
                    ├── sessionManager (cookie-based session auth for dashboard routes)
                    ├── apiauth.RequireMerchantAPIKey (X-API-Key / Bearer / body field auth for merchant API routes)
                    └── handlers/*.go (one file per domain)
```

### Two authentication systems
1. **Session auth** (`backend/libraries/sessionManager/`) — cookie-based, used by all dashboard/admin routes (login, logout, user management). Sessions are in-memory.
2. **Merchant API key auth** (`backend/libraries/apiauth/merchant_api_key.go`) — used for the public widget API and the `/api/v1/merchant/*` programmatic API. The key is accepted from `X-API-Key` header, `Authorization: Bearer`, query param `api_key`, or JSON body field `merchant_api_key`. Keys are stored SHA-256 hashed.

### Route categories (all defined in `backend/libraries/server/handlers/routes.go`)
| Prefix | Auth | Purpose |
|---|---|---|
| `GET /widget/checkout.js` | none | Serves the embeddable checkout widget JS |
| `POST /api/invoices` | Merchant API key | Create invoice (widget flow) |
| `GET /api/invoices/{uuid}` | none | Fetch invoice for checkout display |
| `GET /api/invoices/{uuid}/events` | none | SSE stream for payment status |
| `POST /api/verify` | none | Submit XRPL tx hash for verification |
| `/api/user/*` | session | Auth, signup, merchant membership |
| `/api/dashboard*` | session | Invoice search and analytics |
| `/api/merchant/*` | session | Admin: users, wallet, API keys, webhooks |
| `/api/v1/merchant/customers` | Merchant API key | Programmatic customer CRUD |
| `/api/v1/merchant/invoices` | Merchant API key | Programmatic invoice CRUD |

### XRPL reconciler (`backend/libraries/server/xrpl_reconciler.go`)
Background goroutine that polls the XRP Ledger (XRPL JSON-RPC, defaults to Ripple Testnet) every 5 seconds. For each verified merchant wallet it fetches new transactions since the last checkpoint, matches payments to open invoices (by `destination_tag` first, then XRP amount as fallback), marks matched invoices `paid`, and fires the `invoice.paid` webhook. The ledger cursor is persisted in the `xrpl_checkpoint` table to avoid re-processing.

Environment variable overrides: `XRPL_RPC_URL`, `XRPL_RECONCILE_INTERVAL_SEC`, `XRPL_FALLBACK_USD_PER_XRP`.

### Webhook system (`backend/libraries/webhooks/dispatcher.go`)
`Dispatcher` sends signed `EventEnvelope` payloads to merchant-configured URLs. Signatures use HMAC-SHA256 over `timestamp.body` and are sent in the `X-Webhook-Signature` header as `t=<unix>,v1=<hex>`. Retries up to 3 times with exponential backoff on 408/429/5xx. Every dispatch (success or failure) is written to the `webhook_log` table. Merchants can resend individual log entries via `POST /api/merchant/webhook_logs/{log_id}/resend`.

### Database
- **Development**: SQLite (`backend/xrpay.db`), no process required, auto-created on first run
- **Production**: MySQL, activated by setting `DATABASE_URL` env var
- GORM auto-migrates all models on every startup (additive only — no drops)
- Dev seeding creates three accounts (`dev_admin`, `dev_developer`, `dev_owner`) all with password `password`

### Frontend structure
React 19 + React Router v7. Three React contexts wire the app:
- `AuthContext` — session state and current user
- `MerchantContext` — selected merchant (users can belong to multiple merchants)
- `ThemeContext` — light/dark theme

Pages in `frontend/src/pages/` map 1:1 to routes. Service modules in `frontend/src/services/` encapsulate all API calls.

### Widget
The checkout widget (`backend/static/widget/checkout.js`) is a self-contained JS file embedded in merchant storefronts. It calls `/api/invoices` (create), `/api/invoices/{uuid}` (poll), `/api/invoices/{uuid}/events` (SSE), and `/api/verify`. CORS for these paths is handled by `server.withWidgetCORS` (reflects origin, not `*`). The widget uses Crossmark browser extension for XRPL signing.

## Key environment variables
| Variable | Default | Purpose |
|---|---|---|
| `DATABASE_URL` | — | MySQL DSN; SQLite used if unset |
| `PRODUCTION` | — | Set to `true` for production seed path |
| `XRPL_RPC_URL` | Ripple Testnet | XRPL JSON-RPC endpoint |
| `XRPL_RECONCILE_INTERVAL_SEC` | `5` | Reconciler poll interval |
| `XRPL_FALLBACK_USD_PER_XRP` | `2.0000` | Rate used when Coinbase API is unreachable |
