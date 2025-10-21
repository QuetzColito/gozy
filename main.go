package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Item struct {
	Name      string `json:"name"`
	Color     string `json:"color"`
	Decorator string `json:"decorator"`
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got / request\n")
	io.WriteString(w, "This is my website!\n")
}

func putItems(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	fmt.Printf("got /items/update request\n")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Missing Body", http.StatusBadRequest)
		return
	}
	items := [][]Item{}
	err = json.Unmarshal(body, &items)
	if err != nil {
		http.Error(w, "Body cannot be unmarshalled", http.StatusBadRequest)
		return
	}
	_, err = db.Exec(
		"DELETE FROM LIST_ITEMS")
	if err != nil {
		http.Error(w, "Error Accessing Database", http.StatusInternalServerError)
		return
	}
	for list_index, list := range items {
		for index, item := range list {
			_, err = db.Exec(
				"INSERT INTO LIST_ITEMS (name, index, list, color, decorator) VALUES($1, $2, $3, $4, $5)",
				item.Name, index, list_index, item.Color, item.Decorator)
			if err != nil {
				fmt.Print(err.Error() + "\n")
				http.Error(w, "Error Accessing Database", http.StatusInternalServerError)
				return
			}
		}
	}
	io.WriteString(w, "Successfully updated Items")
}

func getItems(w http.ResponseWriter, db *sql.DB) {
	fmt.Printf("got /items request\n")
	items := [][]Item{}
	for i := range 3 {
		list := []Item{}

		rows, err := db.Query(
			"SELECT name, color, decorator FROM LIST_ITEMS WHERE list = $1 ORDER BY index", i)
		if err != nil {
			http.Error(w, "Error Accessing Database", http.StatusInternalServerError)
			return
		}

		for rows.Next() {
			var item Item
			err = rows.Scan(&item.Name, &item.Color, &item.Decorator)
			list = append(list, item)
		}

		items = append(items, list)
	}

	json, err := json.Marshal(items)
	if err != nil {
		http.Error(w, "Error Accessing Database", http.StatusInternalServerError)
		return
	}
	w.Write(json)
}

func handleItems(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	w.Header().Set("Access-Control-Max-Age", "86400")

	switch r.Method {
	case http.MethodGet:
		getItems(w, db)
	case http.MethodPut:
		putItems(w, r, db)
	case http.MethodOptions:
		http.NoBody.WriteTo(w)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

}

func getEnv() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
}

func connectToDb() *sql.DB {
	psqlInfo := fmt.Sprintf("host=localhost port=5432 user=postgres "+
		"password=%s dbname=cozy sslmode=disable", os.Getenv("POSTGRES_PASSWORD"))
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected!")
	return db
}

func main() {
	getEnv()

	db := connectToDb()
	defer db.Close()
	handleDB := func(pattern string, handler func(w http.ResponseWriter, r *http.Request, Db *sql.DB)) {
		http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) { handler(w, r, db) })
	}

	http.HandleFunc("/", getRoot)
	handleDB("/items", handleItems)

	err := http.ListenAndServe(":3333", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
