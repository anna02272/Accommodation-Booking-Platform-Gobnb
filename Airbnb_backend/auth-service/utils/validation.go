package utils

import "regexp"

func ValidatePassword(password string) bool {
	if !containsUppercase(password) {
		return false
	}

	if !containsLowercase(password) {
		return false
	}

	if !containsDigit(password) {
		return false
	}

	if !containsSpecialCharacter(password) {
		return false
	}

	if len(password) < 8 {
		return false
	}

	return true
}

func containsUppercase(s string) bool {
	return regexp.MustCompile(`[A-Z]`).MatchString(s)
}

func containsLowercase(s string) bool {
	return regexp.MustCompile(`[a-z]`).MatchString(s)
}

func containsDigit(s string) bool {
	return regexp.MustCompile(`\d`).MatchString(s)
}

func containsSpecialCharacter(s string) bool {
	return regexp.MustCompile(`[@$!%*?&.,_]`).MatchString(s)
}
