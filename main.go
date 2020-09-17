package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Book struct (model)
type Book struct {
	ID     string  `json:"id"`
	Isbn   string  `json:"isbn"`
	Title  string  `json:"title"`
	Author *Author `json:"author"`
}

// Author struct
type Author struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

// Init books var as a slice Book struct
var books []Book

// Get all Books from the server
func getBooks(w http.ResponseWriter, r *http.Request) {
	logRequest(r)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

// Get a new Book from the server, given the Book.ID
func getBook(w http.ResponseWriter, r *http.Request) {
	logRequest(r)

	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r) // get request params

	// Loop through books and find with id
	for _, item := range books {
		if item.ID == params["id"] {
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	json.NewEncoder(w).Encode(&Book{})
}

// Add a new Book to the server
func createBook(w http.ResponseWriter, r *http.Request) {
	logRequest(r)

	w.Header().Set("Content-Type", "application/json")

	var book Book
	_ = json.NewDecoder(r.Body).Decode(&book)

	book.ID = strconv.Itoa(rand.Intn(10000000))
	books = append(books, book)

	json.NewEncoder(w).Encode(book)
}

// Update a Book on the server, given the Book.ID
func updateBook(w http.ResponseWriter, r *http.Request) {
	logRequest(r)

	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	// Loop through books and find the ID of book to update
	var book Book
	for i, item := range books {
		if item.ID == params["id"] {
			// similar to deleteBook, cut out the existing version of the book
			books = append(books[:i], books[i+1:]...)

			// add the new version (use existing ID)
			_ = json.NewDecoder(r.Body).Decode(&book)
			book.ID = item.ID
			books = append(books, book)
			break
		}
	}
	// return the updated book
	json.NewEncoder(w).Encode(book)
}

// Delete a Book from the server, given the Book.ID
func deleteBook(w http.ResponseWriter, r *http.Request) {
	logRequest(r)

	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	// Loop through books and find the ID of book to update
	var book Book
	for i, item := range books {
		if item.ID == params["id"] {
			book = item
			// left shift position of all elements to the right of the deleted element by 1
			// i.e., concat [0:i) 	and [i+1,n)
			books = append(books[:i], books[i+1:]...)
			break
		}
	}
	// return the deleted book
	json.NewEncoder(w).Encode(book)
}

// Log any requests to the endpoints
func logRequest(r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Println(err.Error())
	} else {
		log.Println("\n" + string(dump))
	}
}

func main() {
	// Set up some test data
	books = append(books, Book{ID: "1", Isbn: "01", Title: "The Looming Tower", Author: &Author{Firstname: "Lawrence", Lastname: "Wright"}})
	books = append(books, Book{ID: "2", Isbn: "23", Title: "The Sympathiser", Author: &Author{Firstname: "Viet Thanhs", Lastname: "Nguyen"}})
	books = append(books, Book{ID: "3", Isbn: "45", Title: "Catch-22", Author: &Author{Firstname: "Joseph", Lastname: "Wright"}})
	books = append(books, Book{ID: "3", Isbn: "67", Title: "Barbarian Days", Author: &Author{Firstname: "William", Lastname: "Finnegan"}})

	// Initialize router and handler / endpoints
	r := mux.NewRouter()

	r.HandleFunc("/api/books", getBooks).Methods("GET")
	r.HandleFunc("/api/books/{id}", getBook).Methods("GET")
	r.HandleFunc("/api/books", createBook).Methods("POST")
	r.HandleFunc("/api/books/{id}", updateBook).Methods("PUT")
	r.HandleFunc("/api/books/{id}", deleteBook).Methods("DELETE")

	// Create the server
	server := &http.Server{
		Handler:      r,
		Addr:         ":8001",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Configure logging
	LOG_FILE_LOCATION := os.Getenv("LOG_FILE_LOCATION")
	if LOG_FILE_LOCATION != "" {
		logger := &lumberjack.Logger{
			Filename:   LOG_FILE_LOCATION,
			MaxSize:    1,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		}
		log.SetOutput(logger)
	}
	log.Println("Starting book server")
	server.ListenAndServe()
}
