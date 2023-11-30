package utils

import "strconv"

func IsValidInteger(value int) bool {
	// strconv.Itoa returns the string representation of the int
	// If the conversion is successful, and the original string matches the new one,
	// then the value is a valid integer.
	return strconv.Itoa(value) == strconv.FormatInt(int64(value), 10)
}
