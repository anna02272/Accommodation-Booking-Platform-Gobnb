package domain

import "errors"

// specific errors that may occur during the program
// errors.New - error messages
// each function returns some type of error
var (
	errConnectionNotFound      error = errors.New("connection not found")
	errConnectionAlreadyExists error = errors.New("connection already exists")
	errForeignConnection       error = errors.New("not a user's connection")
	errReservationNotFound     error = errors.New("Reservation not found")
	errAccommodationNotFound   error = errors.New("Accommodation not found")
	errUserNotFound            error = errors.New("user not found")
	errInvalidCredentials      error = errors.New("incorrect username or password")
	errInvalidToken            error = errors.New("token invalid")
	errUnauthorized            error = errors.New("unauthorized")
)

func ErrConnectionNotFound() error {
	return errConnectionNotFound
}

func ErrConnectionAlreadyExists() error {
	return errConnectionAlreadyExists
}

func ErrForeignConnection() error {
	return errForeignConnection
}

func ErrReservationNotFound() error {
	return errReservationNotFound
}

func ErrAccommodationNotFound() error {
	return errAccommodationNotFound
}

func ErrUserNotFound() error {
	return errUserNotFound
}

func ErrInvalidCredentials() error {
	return errInvalidCredentials
}

func ErrInvalidToken() error {
	return errInvalidToken
}

func ErrUnauthorized() error {
	return errUnauthorized
}
