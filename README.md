# Crypto Binary Options Backend

A backend service for **crypto binary options trading**, built with **Go**.
The system integrates with **Binance API** for real-time market data, provides **REST APIs** for the frontend, and manages **binary options trades** (UP/DOWN bets on crypto prices within fixed expiry windows).

---

## 📌 Features

* 🔗 **Binance API Integration**

  * Fetches historical OHLC data.
  * Streams live price updates.
* 📈 **Binary Options Trading**

  * Place UP/DOWN trades on crypto pairs.
  * Custom expiry windows (e.g., 1m, 5m).
  * Automatic trade settlement on expiry.
* 🗄 **Database-backed**

  * Stores users, trades, outcomes, and price history.
* 🌐 **RESTful API**

  * Endpoints for history, trades, account management.
* 🔄 **WebSockets**

  * Live price feed pushed to clients.
* 🔒 **Authentication**

  * User accounts & API token system.
* ⚡ **High-performance**

  * Built in Go for scalability & low-latency execution.

---

## 🛠 Tech Stack

* **Language**: Go (1.20+)
* **Database**: PostgreSQL (recommended)
* **API**: REST + WebSocket
* **Data Provider**: Binance API
* **Containerization**: Docker + Docker Compose

---

## ⚙️ Setup

### 1. Clone repo

```bash
git clone https://github.com/your-org/crypto-options-backend.git
cd crypto-options-backend
```

### 2. Configure environment

Create a `.env` file:

```env
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASS=yourpassword
DB_NAME=options_db

BINANCE_API=https://api.binance.com
```

### 3. Run with Docker

```bash
docker-compose up --build
```

Or run locally:

```bash
go mod tidy
go run main.go

or use air 

air
```

---

## 📡 API Endpoints




