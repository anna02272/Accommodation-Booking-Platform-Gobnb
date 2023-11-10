package domain

import "errors"

var (
	errConnectionNotFound      error = errors.New("connection not found")
	errConnectionAlreadyExists error = errors.New("connection already exists")
	errReservationNotFound     error = errors.New("Reservation not found")
	errAccommodationNotFound   error = errors.New("Accommodation not found")
)

// specific errors that may occur during the program
type ReservationError struct {
	Message string
}

func (e ReservationError) Error() string {
	return e.Message
}

func ErrConnectionNotFound() error {
	return errConnectionNotFound
}

func ErrConnectionAlreadyExists() error {
	return errConnectionAlreadyExists
}

func ErrReservationNotFound() error {
	return errReservationNotFound
}

func ErrAccommodationNotFound() error {
	return errAccommodationNotFound
}
