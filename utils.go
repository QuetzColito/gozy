package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func extractBody(w http.ResponseWriter, r *http.Request, target any) bool {
	body, err := io.ReadAll(r.Body)
	if RESOLVE_ERROR_HTTP(err, w, "Missing Body", http.StatusBadRequest) {
		return false
	}

	err = json.Unmarshal(body, target)
	if RESOLVE_ERROR_HTTP(err, w, "Body cannot be unmarshalled", http.StatusBadRequest) {
		return false
	}
	return true
}

func RESOLVE(e error) {
	if e != nil {
		log.Fatal(e.Error())
	}
}

func ERROR_HTTP(w http.ResponseWriter, msg string, status int) {
	log.Print("ERROR: " + msg)
	http.Error(w, msg, status)
}

func RESOLVE_ERROR_HTTP(e error, w http.ResponseWriter, msg string, status int) bool {
	if e == nil {
		return false
	} else {
		log.Print("ERROR: " + e.Error())
		ERROR_HTTP(w, msg, status)
		return true
	}
}

func RESOLVE_AND_INFORM(e error, msg string) {
	if e != nil {
		log.Print(msg)
		log.Fatal(e.Error())
	}
}
