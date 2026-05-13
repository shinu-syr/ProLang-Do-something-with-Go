package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"
)

type Transaction struct {
	ID          int       `json:"id"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Type        string    `json:"type"`
	Category    string    `json:"category"`
	Date        time.Time `json:"date"`
}

type Store struct {
	Transactions []Transaction `json:"transactions"`
	NextID       int           `json:"next_id"`
	mu           sync.Mutex
}

type Summary struct {
	Balance      float64
	TotalIncome  float64
	TotalExpense float64
}

type PageData struct {
	Transactions []Transaction
	Summary      Summary
	Filter       string
}

//Persistence

const dbFile = "finance.json"

func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(dbFile, data, 0644)
}

func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		s.Transactions = []Transaction{}
		s.NextID = 1
		return nil
	}

	data, err := os.ReadFile(dbFile)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, s)
}



func (s *Store) GetSummary() Summary {
	s.mu.Lock()
	defer s.mu.Unlock()

	incomeCh := make(chan float64)
	expenseCh := make(chan float64)

	go func() {
		var total float64
		for _, t := range s.Transactions {
			if t.Type == "income" {
				total += t.Amount
			}
		}
		incomeCh <- total
	}()

	go func() {
		var total float64
		for _, t := range s.Transactions {
			if t.Type == "expense" {
				total += t.Amount
			}
		}
		expenseCh <- total
	}()

	income := <-incomeCh
	expense := <-expenseCh

	return Summary{
		Balance:      income - expense,
		TotalIncome:  income,
		TotalExpense: expense,
	}
}

//Handlers

var (
	store = &Store{}
	tmpl  *template.Template
)

func handleIndex(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	if filter == "" {
		filter = "all"
	}

	summary := store.GetSummary()

	store.mu.Lock()
	filtered := []Transaction{}
	for _, t := range store.Transactions {
		if filter == "all" || t.Type == filter {
			filtered = append(filtered, t)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Date.After(filtered[j].Date)
	})
	store.mu.Unlock()

	data := PageData{
		Transactions: filtered,
		Summary:      summary,
		Filter:       filter,
	}

	err := tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	description := r.FormValue("description")
	amountStr := r.FormValue("amount")
	tType := r.FormValue("type")
	category := r.FormValue("category")

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 || description == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	store.mu.Lock()
	transaction := Transaction{
		ID:          store.NextID,
		Description: description,
		Amount:      amount,
		Type:        tType,
		Category:    category,
		Date:        time.Now(),
	}
	store.Transactions = append(store.Transactions, transaction)
	store.NextID++
	store.mu.Unlock()

	err = store.Save()
	if err != nil {
		log.Printf("Error saving: %v", err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	store.mu.Lock()
	newTransactions := []Transaction{}
	for _, t := range store.Transactions {
		if t.ID != id {
			newTransactions = append(newTransactions, t)
		}
	}
	store.Transactions = newTransactions
	store.mu.Unlock()

	err = store.Save()
	if err != nil {
		log.Printf("Error saving: %v", err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {

	if err := store.Load(); err != nil {
		log.Fatalf("Error loading store: %v", err)
	}


	var err error
	tmpl, err = template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}


	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/add", handleAdd)
	http.HandleFunc("/delete", handleDelete)


	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
