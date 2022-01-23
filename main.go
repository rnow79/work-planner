package main

import (
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// HTTP server port.
const port int = 80

// Verbose bit, for debug.
const verbose bool = true

// Working plan variable. This will be our database.
var workingPlan WorkingPlan

// Main function.
func main() {
	// Look for signing key environ variable.
	var key string = os.Getenv(kName)
	// No environment variable found, inform and stop execution.
	if len(key) == 0 {
		log.Println("Token signing key not found in environment!")
		log.Printf("Its name must be %s. Please create it (b64 encoded).", kName)
		log.Fatalln("Aborting execution.")
	}
	// Decode the key.
	var err error
	signKey, err = base64.StdEncoding.DecodeString(key)
	if err != nil || len(signKey) == 0 {
		log.Fatalln("Environ variable", kName, "must be base64 encoded.")
	}
	// Create router.
	router := mux.NewRouter()
	// Register auth middleware. It will parse and validate the client auth token. If the token is valid
	// it will put the data in the request header and will pass the request to the final endpoint.
	router.Use(authMiddleware)
	// Register static directory.
	router.PathPrefix("/html").Handler(http.StripPrefix("/html/", http.FileServer(http.Dir("./html/"))))
	// Register token generation endpoint.
	router.HandleFunc("/token", postTokenEndpoint).Methods(http.MethodPost)
	// Register workplanner endpoints.
	router.HandleFunc("/", getShiftsEndpoint).Methods(http.MethodGet)
	router.HandleFunc("/", postShiftsEndpoint).Methods(http.MethodPost)
	router.HandleFunc("/", deleteShiftsEndpoint).Methods(http.MethodDelete)
	// Create HTTP server.
	server := &http.Server{Handler: router, Addr: ":" + strconv.Itoa(port), WriteTimeout: 10 * time.Second, ReadTimeout: 10 * time.Second}
	server.ListenAndServe()
}
