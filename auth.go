package main

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

var jwtKey []byte

// TODO: Add User to DB
var users = map[string]string{}

func handleSignin(w http.ResponseWriter, r *http.Request) {
	handleDefault(w, r)

	credentials := Credentials{}
	if !(extractBody(w, r, &credentials)) {
		return
	}

	if users[credentials.Username] != credentials.Password {
		http.Error(w, "Invalid User/Password", http.StatusUnauthorized)
		return
	}

	createTokens(w, r, 1)
}

func createTokens(w http.ResponseWriter, r *http.Request, userId int) {
	now := time.Now()
	refreshExpiresAt := now.Add(48 * time.Hour)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Subject:   string(userId),
		Issuer:    "gozy",
		ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)),
	})
	tokenString, err := token.SignedString(jwtKey)
	if RESOLVE_ERROR_HTTP(err, w, "Error Signing Token", http.StatusInternalServerError) {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Expires:  refreshExpiresAt,
		Secure:   true,
		HttpOnly: true,
	})

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Subject:   string(userId),
		ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
	})
	refreshTokenString, err := refreshToken.SignedString(jwtKey)
	if RESOLVE_ERROR_HTTP(err, w, "Error Signing Token", http.StatusInternalServerError) {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh-token",
		Value:    refreshTokenString,
		Expires:  refreshExpiresAt,
		Secure:   true,
		HttpOnly: true,
	})
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
		createTokens(w, r, 1)
	} else {
		http.Error(w, "Refresh token expired", http.StatusUnauthorized)
		return
	}
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	handleDefault(w, r)
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Expires: time.Now(),
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "refresh-token",
		Expires: time.Now(),
	})

}
