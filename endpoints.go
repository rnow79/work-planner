package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/dgrijalva/jwt-go"
)

// getShifts Endpoint: admins can get all shifts of the week, or a specific user's
// shifts. Workers can just list their own endpoints
func getShiftsEndpoint(w http.ResponseWriter, r *http.Request) {
	if isAdmin(r) {
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
		sendJSON(w, workingPlan.GetUserShifts(getUserIdFromHeader(r)))
		return
	}
}

// postShifts Endpoint: workers can select shifts (if available). Admins can't.
func postShiftsEndpoint(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if isAdmin(r) {
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
			err = workingPlan.InsertUserShift(getUserIdFromHeader(r), day, shift)
			if err != nil {
				sendBadRequest(w, err.Error())
				return
			}
			// If not already present, insert user
			if !workingPlan.HasUser(getUserIdFromHeader(r)) {
				workingPlan.InsertUser(r.Header.Get("X-User"), r.Header.Get("X-Name"), getUserLevelFromHeader(r), getUserIdFromHeader(r))
			}
			// Send current user shifts
			sendJSON(w, workingPlan.GetUserShifts(getUserIdFromHeader(r)))
			return
		}
	}
}

func deleteShiftsEndpoint(w http.ResponseWriter, r *http.Request) {
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
		err = workingPlan.DeleteUserShift(getUserIdFromHeader(r), day, shift)
		if err != nil {
			sendBadRequest(w, err.Error())
			return
		}
	}
}

// Anonymous function for creating tokens with the current server signing key
func getTokenEndpoint(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if usr, nam, lvl, uid := r.PostForm.Get("usr"), r.PostForm.Get("nam"), r.PostForm.Get("lvl"), r.PostForm.Get("uid"); len(usr) == 0 || len(nam) == 0 || len(lvl) == 0 || len(uid) == 0 {
		fmt.Fprintf(w, "error: not enough parameters")
		return
	} else {
		i, err := strconv.Atoi(uid)
		if err != nil {
			fmt.Fprintf(w, "error: uid must be a integer")
			return
		}
		if i < 0 {
			fmt.Fprint(w, "error: uid must be 0 or higher")
			return
		}
		l, err := strconv.Atoi(lvl)
		if err != nil {
			fmt.Fprint(w, "error: lvl must be integer")
			return
		}
		if l != 0 && l != 1 {
			fmt.Fprint(w, "error: level must be 0 or 1")
			return
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"usr": usr, "nam": nam, "uid": i, "lvl": l})
		signed, err := token.SignedString([]byte(signKey))
		if err != nil {
			log.Fatalln("Error signing token")
		}
		fmt.Fprintf(w, signed)
	}
}

// Other functions used by endpoints

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

// Send forbidden
func sendForbidden(w http.ResponseWriter) {
	http.Error(w, "Forbidden", http.StatusForbidden)
}

// Serialize objects in JSON and send
func sendJSON(w http.ResponseWriter, obj interface{}) {
	w.Header().Set("Content-Type", "text/json; charset=utf-8")
	json.NewEncoder(w).Encode(obj)
}

// Determine if user is an admin
func isAdmin(r *http.Request) bool {
	return r.Header.Get("X-Level") == "1"
}

// Extract userid from header and return as an integer
func getUserIdFromHeader(r *http.Request) int {
	i, _ := strconv.Atoi(r.Header.Get("X-Userid"))
	return i
}

// Extract level from header and return as an integer
func getUserLevelFromHeader(r *http.Request) int {
	i, _ := strconv.Atoi(r.Header.Get("X-Level"))
	return i
}
