package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"work-planner/planner"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

// Variables
const port int = 80                 // HTTP server port
const verbose bool = true           // Verbose bit, for debug
const kName string = "SIGNKEY17"    // Signin token key environment variable name
var signKey []byte                  // Key for verify tokens' signature
var workingPlan planner.WorkingPlan // Working plan variable

// Log a line if the program is running in verbose mode
func logLine(format string, v ...interface{}) {
	if verbose {
		log.Println(fmt.Sprintf(format, v...))
	}
}

// Send errors as 400 - Bad Request
func sendBadRequest(w http.ResponseWriter, err string) {
	http.Error(w, err, http.StatusBadRequest)
}

// Serialize objects in JSON and send
func sendJSON(w http.ResponseWriter, obj interface{}) {
	json, _ := json.Marshal(obj)
	w.Header().Set("Content-Type", "text/json; charset=utf-8")
	fmt.Fprintf(w, "%s", string(json))
}

// Parse and Validate token. if valid, fill struct user and return it
func parseToken(token string) (planner.User, error) {
	var returnUser planner.User
	var retErr error = errors.New("error parsing the token")
	if len(signKey) == 0 {
		log.Fatalln("I don't have a signature key") // panic
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

	// Extract claims from interface and return
	returnUser.User, _ = claims["usr"].(string)
	returnUser.Name, _ = claims["nam"].(string)
	lvlstr, _ := claims["lvl"].(string)
	uidstr, _ := claims["uid"].(string)
	returnUser.Level, _ = strconv.Atoi(lvlstr)
	returnUser.UserId, _ = strconv.Atoi(uidstr)
	return returnUser, nil
}

// Default (and unique) endpoint
func shiftHandler(w http.ResponseWriter, r *http.Request) {
	// Parse and validate the token
	httpUser, err := parseToken(r.Header.Get("Auth-Token"))
	if err != nil {
		logLine("Request did not pass the token auth process")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	// Here we have a valid user
	switch r.Method {
	case http.MethodGet:
		if httpUser.IsAdmin() {
			// Admin GET Method
			if userid := r.URL.Query().Get("userid"); len(userid) == 0 {
				// Without params, send all working plan
				sendJSON(w, workingPlan)
				return
			} else {
				userid, err := strconv.Atoi(userid)
				if err != nil {
					sendBadRequest(w, "userid must be numeric")
					return
				}
				// Send specific user shifts
				sendJSON(w, workingPlan.GetUserShifts(userid))
			}
		} else {
			// Worker GET Method
			sendJSON(w, workingPlan.GetUserShifts(httpUser.UserId))
			return
		}
	case http.MethodPost:
		r.ParseForm()
		if httpUser.IsAdmin() {
			// Admin POST Method (Error 400)
			sendBadRequest(w, "admins cannot request shifts")
			return
		} else {
			// Worker POST Method
			if day, shift := r.PostForm.Get("day"), r.PostForm.Get("shift"); len(day) == 0 || len(shift) == 0 {
				sendBadRequest(w, "missing day or shift")
				return
			} else {
				day, err := strconv.Atoi(day)
				if err != nil {
					sendBadRequest(w, "day must be numeric")
					return
				}
				shift, err := strconv.Atoi(shift)
				if err != nil {
					sendBadRequest(w, "shift must be numeric")
					return
				}
				// Try to insert shift
				err = workingPlan.InsertUserShift(httpUser.UserId, day, shift)
				if err != nil {
					sendBadRequest(w, err.Error())
					return
				}
				// If not already present, insert user
				if !workingPlan.HasUser(httpUser.UserId) {
					workingPlan.InsertUser(httpUser)
				}
				// Send current user shifts
				sendJSON(w, workingPlan.GetUserShifts(httpUser.UserId))
				return
			}
		}
	case http.MethodDelete:
		if httpUser.IsAdmin() {
			// Admin DELETE Method
			if userid, day, shift := r.URL.Query().Get("userid"), r.URL.Query().Get("day"), r.URL.Query().Get("shift"); len(userid) == 0 || len(day) == 0 || len(shift) == 0 {
				sendBadRequest(w, "missing day, userid or shift")
				return
			} else {
				userid, err := strconv.Atoi(userid)
				if err != nil {
					sendBadRequest(w, "userid must be numeric")
					return
				}
				day, err := strconv.Atoi(day)
				if err != nil {
					sendBadRequest(w, "day must be numeric")
					return
				}
				shift, err := strconv.Atoi(shift)
				if err != nil {
					sendBadRequest(w, "shift must be numeric")
					return
				}
				// Try to delete user shift
				err = workingPlan.DeleteUserShift(userid, day, shift)
				if err != nil {
					sendBadRequest(w, err.Error())
					return
				}
			}
		} else {
			// Worker DELETE Method
			if day, shift := r.URL.Query().Get("day"), r.URL.Query().Get("shift"); len(day) == 0 || len(shift) == 0 {
				sendBadRequest(w, "missing day or shift")
				return
			} else {
				day, err := strconv.Atoi(day)
				if err != nil {
					sendBadRequest(w, "day must be numeric")
					return
				}
				shift, err := strconv.Atoi(shift)
				if err != nil {
					sendBadRequest(w, "shift must be numeric")
					return
				}
				// Try to delete user shift
				err = workingPlan.DeleteUserShift(httpUser.UserId, day, shift)
				if err != nil {
					sendBadRequest(w, err.Error())
					return
				}
			}
		}
	}
}

func main() {
	// Look for signing key environ variable
	keyFromEnv := os.Getenv(kName)

	// No environment variable found, inform and stop execution
	if len(keyFromEnv) == 0 {
		log.Println("Token signing key not found in environment!")
		log.Printf("Its name must be %s. Please create it (b64 encoded).", kName)
		log.Fatalln("Aborting execution.")
	}

	// Decode the key
	var err error
	signKey, err = base64.StdEncoding.DecodeString(keyFromEnv)
	if err != nil || len(signKey) == 0 {
		log.Fatalln("Environ variable", kName, "must be base64 encoded.")
	}
	// Create router
	router := mux.NewRouter()
	router.HandleFunc("/", shiftHandler).Methods(http.MethodGet, http.MethodPost, http.MethodDelete)
	// Create HTTP server
	server := &http.Server{Handler: router, Addr: ":" + strconv.Itoa(port), WriteTimeout: 10 * time.Second, ReadTimeout: 10 * time.Second}
	server.ListenAndServe()
}
