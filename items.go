package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Item struct {
	Name      string `json:"name"`
	Color     string `json:"color"`
	Decorator string `json:"decorator"`
}

func putItems(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	body, err := io.ReadAll(r.Body)
	if RESOLVE_ERROR_HTTP(err, w, "Missing Body", http.StatusBadRequest) {
		return
	}

	items := [][]Item{}
	err = json.Unmarshal(body, &items)
	if RESOLVE_ERROR_HTTP(err, w, "Body cannot be unmarshalled", http.StatusBadRequest) {
		return
	}

	_, err = db.Exec(
		"DELETE FROM list_items")
	if RESOLVE_ERROR_HTTP(err, w, "Error Accessing Database", http.StatusInternalServerError) {
		return
	}

	for list_index, list := range items {
		for index, item := range list {
			_, err = db.Exec(
				"INSERT INTO list_items (name, index, list, color, decorator) VALUES($1, $2, $3, $4, $5)",
				item.Name, index, list_index, item.Color, item.Decorator)
			if RESOLVE_ERROR_HTTP(err, w, "Error Accessing Database", http.StatusInternalServerError) {
				return
			}
		}
	}
	io.WriteString(w, "Successfully updated Items")
}

func getItems(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	fmt.Printf("got /items request\n")
	items := [][]Item{}
	for i := range 3 {
		list := []Item{}

		rows, err := db.Query(
			"SELECT name, color, decorator FROM list_items WHERE list = $1 ORDER BY index", i)
		if err != nil {
			http.Error(w, "Error Accessing Database", http.StatusInternalServerError)
			return
		}

		for rows.Next() {
			var item Item
			err = rows.Scan(&item.Name, &item.Color, &item.Decorator)
			if err != nil {
				http.Error(w, "Error Accessing Database", http.StatusInternalServerError)
				return
			}
			list = append(list, item)
		}

		items = append(items, list)
	}

	json, err := json.Marshal(items)
	if err != nil {
		http.Error(w, "Could not Marshal Items", http.StatusInternalServerError)
		return
	}
	w.Write(json)
}

func handleItems(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	switch r.Method {
	case http.MethodGet:
		getItems(w, r, db)
	case http.MethodPut:
		putItems(w, r, db)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
