package main

import "errors"

// Amount of shifts per day.
const shiftsPerDay int = 3

// Maximum shifts a user is allowed to own per day.
const maxShiftsPerUserPerDay int = 1

// Maximum shifts a user is allowed to own per week.
const maxShiftsPerUserPerWeek int = 5

// User struct.
type User struct {
	User   string `json:"user"`   // Username (ex john)
	Name   string `json:"name"`   // Friendly name (ex John Smith).
	Level  int    `json:"level"`  // User level: 0 means worker, 1 means admin.
	UserId int    `json:"userid"` // User id.
}

// Shift struct
type Shift struct {
	Used   bool `json:"used"`   // Shift is used or not.
	Userid int  `json:"userid"` // User id.
}

// Day struct
type Day struct {
	Shifts [shiftsPerDay]Shift `json:"shifts"` // Each day contains 3 shifts.
}

// UserShift struct: used with UserShifts.
type UserShift struct {
	Day   int `json:"day"`   // Day of the shift.
	Shift int `json:"shift"` // Numer of shift.
}

// UserShifts struct: Used for return all shifts of a particular user.
type UserShifts struct {
	Userid int         `json:"userid"` // User id.
	Shifts []UserShift `json:"shifts"` // Array of shifts.
}

// WorkingPlan struct.
type WorkingPlan struct {
	Users []User `json:"users"` // Array of users who have shifts.
	Days  [7]Day `json:"days"`  // Seven days of the week.
}

// Checks if the workplan has information about a user.
func (w *WorkingPlan) HasUser(userid int) bool {
	for _, user := range w.Users {
		if user.UserId == userid {
			return true
		}
	}
	return false
}

// Inserts user data in the workplan.
func (w *WorkingPlan) InsertUser(user string, name string, level int, userid int) {
	if !w.HasUser(userid) {
		newUser := &User{user, name, level, userid}
		w.Users = append(w.Users, *newUser)
	}
}

// Deletes user data from the workplan.
func (w *WorkingPlan) deleteUser(userid int) {
	for index, user := range w.Users {
		if user.UserId == userid {
			w.Users = append(w.Users[:index], w.Users[index+1:]...)
			return
		}
	}
}

// Tells the number of shifts some user has in the entire week.
func (w *WorkingPlan) getUserShifCount(userid int) (count int) {
	for _, day := range w.Days {
		for _, shift := range day.Shifts {
			if shift.Used && shift.Userid == userid {
				count++
			}
		}
	}
	return
}

// Gets user shifts.
func (w *WorkingPlan) GetUserShifts(userid int) (userShifts UserShifts) {
	userShifts.Userid = userid
	for iday, day := range w.Days {
		for ishift, shift := range day.Shifts {
			if shift.Used && shift.Userid == userid {
				sh := &UserShift{iday, ishift}
				userShifts.Shifts = append(userShifts.Shifts, *sh)
			}
		}
	}
	return
}

// Gets the shifts of an user a specific day.
func (w *WorkingPlan) GetUserShiftsCountByDay(userid int, day int) (shifts int) {
	if !IsValidDay(day) {
		return
	}
	for _, shift := range w.Days[day].Shifts {
		if shift.Userid == userid {
			shifts++
		}
	}
	return
}

// Inserts a shift.
func (w *WorkingPlan) InsertUserShift(userid int, day int, shift int) error {
	if !IsValidDay(day) {
		return errors.New("invalid day")
	}
	if !IsValidShift(shift) {
		return errors.New("invalid shift")
	}
	if w.GetUserShiftsCountByDay(userid, day) >= maxShiftsPerUserPerDay {
		return errors.New("max shifts per day reached")
	}
	if w.getUserShifCount(userid) >= maxShiftsPerUserPerWeek {
		return errors.New("max shifts per week reached")
	}
	if w.Days[day].Shifts[shift].Used {
		return errors.New("shift unavailable")
	}
	w.Days[day].Shifts[shift].Used = true
	w.Days[day].Shifts[shift].Userid = userid
	return nil
}

// Deletes a shift.
func (w *WorkingPlan) DeleteUserShift(userid int, day int, shift int) error {
	if !IsValidDay(day) {
		return errors.New("invalid day")
	}
	if !IsValidShift(shift) {
		return errors.New("invalid shift")
	}
	if w.Days[day].Shifts[shift].Used && w.Days[day].Shifts[shift].Userid == userid {
		w.Days[day].Shifts[shift].Used = false
		w.Days[day].Shifts[shift].Userid = 0
	} else {
		return errors.New("user does not own the shift")
	}
	// If user has no remaining shifts in the planner, delete his data from planner.
	if w.getUserShifCount(userid) == 0 {
		w.deleteUser(userid)
	}
	return nil
}

// Valid days are from 0 to 6.
func IsValidDay(day int) bool {
	if day >= 0 && day < 7 {
		return true
	}
	return false
}

// Valid shifts are from 0 to shiftsPerDay constant.
func IsValidShift(shift int) bool {
	if shift >= 0 && shift < shiftsPerDay {
		return true
	}
	return false
}
