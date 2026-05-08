# Finance Manager — Go Web App Design Spec

## Overview
Build a single-page personal finance manager web app using **Go (Golang)** with zero external dependencies. The entire app lives in one `main.go` file. The UI is served by Go's built-in `net/http` package and rendered via `html/template`. Data is persisted to a local `finance.json` file using `encoding/json`.

---

## Stack & Constraints
- **Language:** Go (1.21+)
- **No external packages** — only Go standard library
- **Single file:** `main.go` only (plus auto-generated `finance.json`)
- **No frontend frameworks** — plain HTML, CSS, vanilla JS inside the Go template
- **Server runs at:** `http://localhost:8080`

---

## Data Model

### Transaction
```go
type Transaction struct {
    ID          int       `json:"id"`
    Description string    `json:"description"`
    Amount      float64   `json:"amount"`
    Type        string    `json:"type"`     // "income" or "expense"
    Category    string    `json:"category"` // e.g. "Food", "Salary", "Rent", "Transport", "Other"
    Date        time.Time `json:"date"`
}
```

### Store
```go
type Store struct {
    Transactions []Transaction `json:"transactions"`
    NextID       int           `json:"next_id"`
}
```

### Summary (computed, not stored)
```go
type Summary struct {
    Balance      float64
    TotalIncome  float64
    TotalExpense float64
}
```

---

## Features

### 1. Add Transaction
- Form fields:
  - `description` — text input, required
  - `amount` — number input, required, min 0.01, step 0.01
  - `type` — dropdown: `Income` / `Expense`
  - `category` — dropdown: `Allowance`, `School expense`, `Food`, `Rent`, `Transportation`, `Utilities`, `Entertainment`, `Activities`, `Other`
- On submit: POST to `/add`, save to JSON, redirect back to `/`
- Validate: amount must be > 0, description must not be empty

### 2. Live Balance Summary
- Three cards always visible at the top of the page:
  - **Balance** = Total Income − Total Expenses
  - **Total Income** = sum of all income transactions
  - **Total Expenses** = sum of all expense transactions
- Computed fresh on every page load using **goroutines + channels** (one goroutine for income, one for expenses, results sent back via channels)

### 3. Transaction Table
- Shows all transactions in reverse chronological order (newest first)
- Columns: `#`, `Description`, `Category`, `Type`, `Amount`, `Date`, `Action`
- Type column shows a colored badge: green for Income, red for Expense
- Amount is formatted with 2 decimal places and a `₱` peso sign
- Date is formatted as `Jan 02, 2006`

### 4. Delete Transaction
- Each row has a Delete button
- POST to `/delete` with the transaction ID
- Removes from slice, saves updated JSON, redirects to `/`

### 5. Filter by Type
- Three filter buttons above the table: `All`, `Income`, `Expense`
- Implemented as GET query params: `/?filter=income`, `/?filter=expense`, `/?filter=all`
- Active filter button is visually highlighted
- Filtering happens server-side in the handler

### 6. Data Persistence
- On every Add and Delete, save the full store to `finance.json` using `json.MarshalIndent`
- On server start, load from `finance.json` if it exists; otherwise start fresh with `NextID: 1`
- Use `sync.Mutex` on all read/write operations to the store

---

## HTTP Routes

| Method | Route     | Description                        |
|--------|-----------|------------------------------------|
| GET    | `/`       | Render main page, optional `?filter=` param |
| POST   | `/add`    | Add a new transaction              |
| POST   | `/delete` | Delete transaction by ID           |

---

## UI Design

### Theme
- **Dark mode only**
- Background: `#0f1117`
- Card/panel background: `#1e2231`
- Border color: `#2d3348`
- Primary text: `#e2e8f0`
- Muted text: `#64748b`
- Income accent: `#76D675` (green)
- Expense accent: `#ED6C6A` (red)
- Balance accent: `#74B6F9` (blue)
- Button primary: `#59916b`

### Typography
- Font: `'Inter', system-ui, sans-serif`
- Page title: centered, `1.8rem`, letter-spacing



### Layout (top to bottom, single page)
1. **Header** — App title `Rhic's DigiBaon` + subtitle 
2. **Summary Cards** — Three equal cards in a flex row (Balance, Income, Expenses), full width
3. **Bottom Section** — Two columns side by side:
   - **Left (box 1): Transaction Table** — takes ~65% of the width, shows all transactions with filter buttons above it
   - **Right (box 2): Add Transaction Form** — takes ~35% of the width, a contained card with all input fields stacked vertically and a submit button at the bottom
4. **Empty State** — Shown inside the table area if no transactions match the current filter

### Rough Layout
```
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│   BALANCE   │  │   INCOME    │  │   EXPENSES  │
└─────────────┘  └─────────────┘  └─────────────┘

┌──────────────────────────────┐  ┌───────────────┐
│  [ All ] [ Income ] [Expense]│  │ Add Transaction│
│  ┌────────────────────────┐  │  │               │
│  │ # │ Desc │ Type │ Amt  │  │  │ Description   │
│  │───┼──────┼──────┼──────│  │  │ [___________] │
│  │ 1 │ ...  │ INC  │ ₱... │  │  │               │
│  │ 2 │ ...  │ EXP  │ ₱... │  │  │ Amount        │
│  └────────────────────────┘  │  │ [___________] │
└──────────────────────────────┘  │               │
                                  │ Type          │
                                  │ [___________] │
                                  │               │
                                  │ Category      │
                                  │ [___________] │
                                  │               │
                                  │ [ + Add ]     │
                                  └───────────────┘
```



---

## Go Concurrency Requirements
The `GetSummary()` function **must** use goroutines and channels as follows:
```go
func (s *Store) GetSummary() Summary {
    incomeCh := make(chan float64)
    expenseCh := make(chan float64)

    go func() {
        // sum all income transactions
        incomeCh <- total
    }()

    go func() {
        // sum all expense transactions
        expenseCh <- total
    }()

    income := <-incomeCh
    expense := <-expenseCh
    // compute and return Summary
}
```
This is required to demonstrate Go's concurrency features for the activity.

---

## File Structure
```
financeapp/
├── main.go        ← entire app (server, handlers, templates, styles)
├── go.mod         ← module: financeapp, go 1.22
└── finance.json   ← auto-created on first transaction
```

---

## How to Run
```bash
go mod init financeapp
go run main.go
# Open http://localhost:8080
```
