package main

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

var jwtKey []byte

const (
	JWT_COOKIE     = "token"
	REFRESH_COOKIE = "refreshToken"
)

// extracts the currently logged in User from the cookies.
// Refreshes JWT if Necessary and Possible
func getLoggedInUserId(w http.ResponseWriter, r *http.Request, db *sql.DB) (int, error) {
	c, err := r.Cookie(JWT_COOKIE)
	needsTokenRefresh := false
	if err != nil {
		c, err = r.Cookie(REFRESH_COOKIE)
		if err != nil {
			return -1, err
		}
		needsTokenRefresh = true
	}

	tkn, err := jwt.Parse(c.Value, func(token *jwt.Token) (any, error) {
		return jwtKey, nil
	})
	if !tkn.Valid || err != nil {
		http.Error(w, "Invalid Token", http.StatusBadRequest)
		return -1, err
	}

	userIdString, err := tkn.Claims.GetSubject()
	if err != nil {
		http.Error(w, "Token missing Subject", http.StatusBadRequest)
		return -1, err
	}

	id, err := strconv.Atoi(userIdString)
	if err != nil {
		http.Error(w, "Subject is not a number", http.StatusBadRequest)
		return -1, err
	}

	if needsTokenRefresh && userExists(w, db, id) {
		createTokens(w, id)
	}

	return id, nil
}

// Checks if a User is still present in the Database
func userExists(w http.ResponseWriter, db *sql.DB, userId int) bool {
	row := db.QueryRow("SELECT id FROM users WHERE id = $1;", userId)
	var id int
	err := row.Scan(&id)
	if RESOLVE_ERROR_HTTP(err, w, "Error retrieving User", http.StatusUnauthorized) {
		return false
	}
	return true
}

func setToken(w http.ResponseWriter, userId int, name string, expiresAt time.Time) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Subject:   strconv.Itoa(userId),
		Issuer:    "gozy",
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	})
	tokenString, err := token.SignedString(jwtKey)
	if RESOLVE_ERROR_HTTP(err, w, "Error Signing Token", http.StatusInternalServerError) {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    tokenString,
		Expires:  expiresAt,
		Secure:   true,
		HttpOnly: true,
		Path:     "/",
	})
}

func createTokens(w http.ResponseWriter, userId int) {
	now := time.Now()

	setToken(w, userId, "token", now.Add(5*time.Minute))
	setToken(w, userId, "refreshToken", now.Add(48*time.Hour))
}

func handleRefresh(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("token")
	if err == nil {
		http.Error(w, "Still Authenticated", http.StatusBadRequest)
		return
	}

	c, err := r.Cookie("refresh-token")
	if err != nil {
		http.Error(w, "Missing Refresh Token", http.StatusUnauthorized)
		return
	}

	tkn, err := jwt.Parse(c.Value, func(token *jwt.Token) (any, error) {
		return jwtKey, nil
	})
	if !tkn.Valid {
		http.Error(w, "Invalid Refresh Token", http.StatusBadRequest)
		return
	}
	if RESOLVE_ERROR_HTTP(err, w, "Invalid Refresh Token", http.StatusBadRequest) {
		return
	}

	expiresAt, err := tkn.Claims.GetExpirationTime()
	if RESOLVE_ERROR_HTTP(err, w, "Invalid Refresh Token", http.StatusBadRequest) {
		return
	}

	if time.Until(expiresAt.Time) > 0 {
		createTokens(w, 1)
	} else {
		http.Error(w, "Refresh token expired", http.StatusUnauthorized)
		return
	}
}

// Request is unused but kept as an Argument to Match the typical handle-function signature
// Unsets the Login Cookies
func handleLogout(w http.ResponseWriter, r *http.Request) {
	setToken(w, -1, "token", time.Unix(0, 0))
	setToken(w, -1, "refreshToken", time.Unix(0, 0))
}
