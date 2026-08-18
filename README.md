# 🎟️ IT Security & Identity Management Ticket Analytics Dashboard

A modern, high-performance web application designed for processing, analyzing, and visualizing IT Security & Identity Management support tickets. The platform enables automated Excel report ingestion, major incident auto-detection, performance analytics, and operational metrics tracking.

---

## 🚀 Tech Stack

### **Backend**
- **Language & Engine**: Go `1.25+`
- **Web Framework**: [Gin Gonic](https://github.com/gin-gonic/gin)
- **ORM & Database**: [GORM](https://gorm.io/) with [SQLite3](https://github.com/mattn/go-sqlite3)
- **Excel Parser**: [Excelize v2](https://github.com/xuri/excelize)

### **Frontend**
- **Framework**: [SvelteKit 5](https://kit.svelte.dev/)
- **Build Tool**: [Vite](https://vitejs.dev/)
- **Styling**: [Tailwind CSS v4](https://tailwindcss.com/)
- **Language**: TypeScript

---

## ✨ Key Features

- 📊 **Interactive Dashboard (`/`)**: High-level overview of ticket metrics, search filtering, monthly scope controls, and raw ticket table inspection.
- 📈 **Advanced Category & Domain Analytics (`/analysis`)**: In-depth distribution metrics across product classification tiers (Tier 1, Tier 2, Tier 3), sites, priority levels, and status categories.
- 🚨 **Major Incident Tracker (`/incident`)**: Automated detection of potential incidents (flagging days with $\ge 50$ tickets), root cause documentation, and historical incident logs.
- 🏆 **Team & Assignee Leaderboard (`/leaderboard`)**: Individual assignee performance, team workload distribution, and ticket resolution statistics.
- 📁 **Seamless Excel File Ingestion**: Instant batch import of standard `.xlsx` ticket dump files with built-in schema validation and date parsing (`Asia/Bangkok` timezone support).

---

## 📂 Project Structure

```text
dashboard_ticket/
├── backend/
│   ├── main.go               # Gin HTTP server & route definitions
│   ├── handlers.go           # API handlers (upload, tickets, incidents, analytics)
│   ├── models.go             # GORM models (Ticket, Incident)
│   ├── parser.go             # Excel parsing logic using Excelize
│   ├── tickets.db            # SQLite database file
│   └── ticket_excel_format.xlsx # Sample input Excel template
├── frontend/
│   ├── src/
│   │   ├── routes/
│   │   │   ├── +page.svelte           # Overview Dashboard
│   │   │   ├── analysis/+page.svelte  # Detailed Analytics Page
│   │   │   ├── incident/+page.svelte  # Major Incident Page
│   │   │   └── leaderboard/+page.svelte # Assignee Leaderboard
│   │   ├── app.css
│   │   └── app.html
│   ├── package.json
│   ├── svelte.config.js
│   └── vite.config.ts
└── README.md
```

---

## 🛠️ Installation & Setup

### **Prerequisites**
- [Go](https://go.dev/doc/install) `1.25` or higher
- [Node.js](https://nodejs.org/) `v18+` and `npm`
- `gcc` / C compiler (required by CGO for SQLite driver)

---

### 1. Run Backend Server

```bash
cd backend

# Install Go dependencies
go mod download

# Run the backend application
go run .
# Or compile and execute the binary:
# go build -o ticket-dashboard . && ./ticket-dashboard
```

The backend server runs on `http://localhost:8081`.

---

### 2. Run Frontend Development Server

```bash
cd frontend

# Install dependencies
npm install

# Start Vite dev server
npm run dev
```

Open your browser at `http://localhost:5173` (or the URL displayed in your terminal).

---

## 🔌 API Reference

| Endpoint | Method | Description |
|---|---|---|
| `/upload` | `POST` | Upload `.xlsx` file (`multipart/form-data` with field `file`) |
| `/months` | `GET` | List distinct ticket creation months (`YYYY-MM`) |
| `/tickets` | `GET` | Query tickets. Supports `?month=YYYY-MM` parameter |
| `/incidents` | `GET` | Retrieve list of recorded major incidents |
| `/incidents` | `POST` | Upsert major incident details (`date`, `title`, `description`, `root_cause`, `ticket_count`) |
| `/detected-incidents` | `GET` | Returns dates where total ticket volume $\ge 50$ |

---

## 📑 Supported Excel File Format

The parser expects `.xlsx` files with a header row containing the following fields:

| Field Name | Status | Example |
|---|---|---|
| `Ticket ID` | **Required** | `INC1234567` |
| `Ticket Type` | **Required** | `User Request` / `Incident` |
| `Priority` | **Required** | `Medium` / `High` / `Critical` |
| `Created Date` | **Required** | `06/07/2026 5:25 PM` |
| `Subject` | Optional | `Password Reset` |
| `Domain Group` | Optional | `Identity Management` |
| `Company` | Optional | `Company A` |
| `Country` | Optional | `Thailand` |
| `Assignee` | Optional | `Staff Name` |
| `Group Assignee` | Optional | `THL2 IT Security Identity Mgmt` |
| `Product Categorization Tier 1` | Optional | `Identity Management` |
| `Product Categorization Tier 2` | Optional | `User Account Access` |
| `Product Categorization Tier 3` | Optional | `Onelogin` |
| `Resolved Date` | Optional | `07/07/2026 9:28 AM` |
| `Closed Date` | Optional | `08/07/2026 10:00 AM` |
| `Completed Date` | Optional | `07/07/2026 9:30 AM` |
