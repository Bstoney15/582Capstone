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

## Seeded Development Accounts

When running the database in development mode, the following seeded accounts are available for immediate use. 

All accounts use the same password: `password`

| Role | Username |
| :--- | :--- |
| **Admin** | `dev_admin` |
| **Developer** | `dev_developer` |
| **Owner** | `dev_owner` |
