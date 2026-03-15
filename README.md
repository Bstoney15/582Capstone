# XRPay

**XRPay** is a platform as a service that orchestrates XRP payments for online merchants, allowing them to easily accept payment in XRP. 

## Built With
- **Frontend:** React + Vite
- **Backend:** Golang
- **DB:** 
    - Default (development): uses SQLite locally with xrpay.db file, no separate DB process required
    - Production: When environment variable `production=true`, the backend tries to use MySQL

## Steps to run

### Prerequisites
- npm
- Go
- MySQL(if running in production)

### Running Backend

From project root, run:
```
cd backend
go mod tidy # one time, just to install dependencies
go run ./main.go # starts backend on port 8080
```
Now you have the backend server running on port 8080.

### Running Frontend

In a new terminal session (do not close or stop backend process to run), run:
```
cd frontend
npm install # one time, just to install dependencies
npm run dev # starts dev server at localhost:5173
```

Now that you have both the frontend and backend servers running, you can open up the frontend URL in your browser (`http://localhost:5173`) and interact with the XRPay webpage.

## Checkout Widget (MVP)

The backend now serves a drop-in checkout widget at:

`http://localhost:8080/widget/checkout.js`

Example merchant page usage:

```html
<button id="pay-with-xrpl">Pay with XRP</button>

<script src="http://localhost:8080/widget/checkout.js"></script>
<script>
    MyPay.init({
        invoiceId: "replace-with-invoice-uuid",
        triggerSelector: "#pay-with-xrpl",
        successUrl: "https://merchant.example/success",
        apiBaseUrl: "http://localhost:8080"
    });
</script>
```

`MyPay.init(...)` creates the modal immediately, but payment flow starts only when the merchant-provided trigger element is clicked.
The widget is Crossmark-only for now and shows a status warning if the extension is not available.

### Widget API Dependencies (MVP)

The widget calls these backend endpoints:

- `POST /api/invoices`
    - Request body: `{ "merchant_api_key": "dev_demo_invoice_key", "amount_usd": "12.34" }` (or `amount_xrp`)
    - Returns `201 Created` with `{ "invoice_id": "...", "amount_xrp": "...", "usd_per_xrp": "..." }`
    - Dev-only: API key is intentionally hardcoded for local storefront demo
- `GET /api/invoices/{uuid}`
    - Returns `invoiceId`, `amountDrops`, `destinationTag`, `merchantAddress`
- `POST /api/verify`
    - Request body: `{ "invoice_id": "...", "tx_hash": "..." }`
    - `tx_hash` may be an XRPL tx hash (64 hex) or a Crossmark payload/request id (UUID)
    - Returns `202 Accepted` when verification is queued

### Frontend Demo Page

With backend and frontend both running, open:

`http://localhost:5173/widget-demo`

This page loads `http://localhost:8080/widget/checkout.js`, initializes `MyPay`, and binds to a demo button using `triggerSelector`.

## Seeded Development Accounts

When running the database in development mode, the following seeded accounts are available for immediate use. 

All accounts use the same password: `password`

| Role | Username |
| :--- | :--- |
| **Admin** | `dev_admin` |
| **Developer** | `dev_developer` |
| **Owner** | `dev_owner` |
