package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

// Key for verify tokens' signature.
var signKey []byte

// Signin token key environment variable name.
const kName string = "SIGNKEY17"

// Token header name.
const headerName string = "Auth-Token"

// Each request must include a header with a valid token, otherwise a forbidden response is sent.
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Exclude static directory and token generation api.
		if strings.HasPrefix(strings.ToLower(r.RequestURI), "/html") || strings.HasPrefix(strings.ToLower(r.RequestURI), "/token") {
			next.ServeHTTP(w, r)
			return
		}
		// Extract user data from token.
		user, err := extractToken(r.Header.Get(headerName))
		if err != nil {
			logLine("Error extracting token: %s", err)
			sendForbidden(w)
			return
		}
		// Insert user data in request headers.
		r.Header.Set("X-User", user.User)
		r.Header.Set("X-Name", user.Name)
		r.Header.Set("X-Level", strconv.Itoa(user.Level))
		r.Header.Set("X-Userid", strconv.Itoa(user.UserId))
		// Call the endpoint.
		next.ServeHTTP(w, r)
	})
}

// Parse and validate token. If valid, fill struct user and return it.
func extractToken(token string) (User, error) {
	var returnUser User
	var retErr error = errors.New("error parsing the token")
	if len(signKey) == 0 {
		log.Fatalln("I don't have a signature key") // panic!
	}
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return signKey, nil
	})
	if err != nil {
		logLine("error processing the token: %s", err)
		return returnUser, retErr
	}
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || !parsedToken.Valid {
		logLine("error validating the token")
		return returnUser, retErr
	}
	// Extract claims from interface and return.
	returnUser.User, _ = claims["usr"].(string)
	returnUser.Name, _ = claims["nam"].(string)
	lvlstr, _ := claims["lvl"].(string)
	uidstr, _ := claims["uid"].(string)
	returnUser.Level, _ = strconv.Atoi(lvlstr)
	returnUser.UserId, _ = strconv.Atoi(uidstr)
	return returnUser, nil
}
