package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

type User struct {
	Name string `json:"name"`
}

// Handler for /user Endpoint
func handleUser(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	switch r.Method {
	case http.MethodGet:
		getUser(w, r, db)
	case http.MethodPut:
		createUser(w, r, db)
	case http.MethodPost:
		loginUser(w, r, db)
	case http.MethodDelete:
		deleteUser(w, r, db)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// gets the currently Logged in User, or null if not logged in
func getUser(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id, err := getLoggedInUserId(w, r, db)
	if err != nil {
		json, _ := json.Marshal(nil)
		w.Write(json)
		return
	}

	var user User
	row := db.QueryRow("SELECT name FROM users WHERE id = $1", id)
	err = row.Scan(&user.Name)
	if RESOLVE_ERROR_HTTP(err, w, "Could not get User Details", http.StatusInternalServerError) {
		return
	}

	json, err := json.Marshal(user)
	if RESOLVE_ERROR_HTTP(err, w, "Could not Marshal User", http.StatusInternalServerError) {
		return
	}

	w.Write(json)
}

func loginUser(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	credentials := Credentials{}
	if !(extractBody(w, r, &credentials)) {
		return
	}

	rows := db.QueryRow("SELECT id FROM users WHERE name = $2 AND crypt($1, password) = password;", credentials.Password, credentials.Username)
	var id int
	err := rows.Scan(&id)
	if RESOLVE_ERROR_HTTP(err, w, "Invalid User/Password", http.StatusUnauthorized) {
		return
	}
	log.Printf("logged in User %s, (ID: %d)", credentials.Username, id)

	createTokens(w, id)
}

// Creates a User and Automatically logs in as that user (might change that in the future)
func createUser(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	credentials := Credentials{}
	if !(extractBody(w, r, &credentials)) {
		return
	}

	_, err := db.Exec("INSERT INTO users (name, password) VALUES ($1, crypt($2, gen_salt('md5')))", credentials.Username, credentials.Password)
	if RESOLVE_ERROR_HTTP(err, w, "Unable to insert User into DB", http.StatusBadRequest) {
		return
	}

	rows, err := db.Query("SELECT id FROM users WHERE name = $1", credentials.Username)
	if RESOLVE_ERROR_HTTP(err, w, "Couldn't find created User", http.StatusBadRequest) {
		return
	}

	var id int
	for rows.Next() {
		err = rows.Scan(&id)
		if RESOLVE_ERROR_HTTP(err, w, "Couldn't find created User", http.StatusBadRequest) {
			return
		}
	}
	log.Printf("created User %s, (ID: %d)", credentials.Username, id)

	createTokens(w, id)
}

// Deletes the currently logged in User and does a Logout
func deleteUser(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id, err := getLoggedInUserId(w, r, db)
	if err != nil {
		http.Error(w, "Missing Login", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM users WHERE id = $1", id)
	if RESOLVE_ERROR_HTTP(err, w, "Unable to Delete User from DB", http.StatusBadRequest) {
		return
	}
	log.Printf("deleted User %d", id)

	handleLogout(w, r)
}
