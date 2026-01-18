package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func getRoot(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "This is the backend for quetz.dev!\n")
}

func handleDefault(w http.ResponseWriter, r *http.Request) bool {
	log.Printf("got %s request", r.Pattern)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	w.Header().Set("Access-Control-Max-Age", "86400")

	if r.Method == http.MethodOptions {
		http.NoBody.WriteTo(w)
		return true
	} else {
		return false
	}
}

func getEnv() {
	f, err := os.Open(".env")
	RESOLVE_AND_INFORM(err, "Error reading .env file")

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		split := strings.SplitN(sc.Text(), "=", 2)
		if len(split) == 2 {
			os.Setenv(split[0], split[1])
		}
	}
	jwtKey = []byte("OAUTH_SECRET")
}

func connectToDb() *sql.DB {
	getEnv()
	psqlInfo := fmt.Sprintf("host=localhost port=5432 user=postgres "+
		"password=%s dbname=cozy sslmode=disable", os.Getenv("POSTGRES_PASSWORD"))
	db, err := sql.Open("postgres", psqlInfo)
	RESOLVE_AND_INFORM(err, "Could not open Connection to DB")
	err = db.Ping()
	RESOLVE_AND_INFORM(err, "Could not connect to DB")
	log.Print("Successfully connected!")
	return db
}

func main() {
	db := connectToDb()
	defer db.Close()
	handleDB := func(pattern string, handler func(w http.ResponseWriter, r *http.Request, Db *sql.DB)) {
		http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			if handleDefault(w, r) {
				return
			}
			handler(w, r, db)
		})
	}
	handle := func(pattern string, handler func(w http.ResponseWriter, r *http.Request)) {
		http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			if handleDefault(w, r) {
				return
			}
			handler(w, r)
		})
	}

	handle("/", getRoot)
	// handleDB("/login", handleLogin)
	handle("/logout", handleLogout)
	handleDB("/user", handleUser)
	handle("/refresh", handleRefresh)
	handleDB("/items", handleItems)

	log.Fatal(http.ListenAndServe(":3333", nil))
}
