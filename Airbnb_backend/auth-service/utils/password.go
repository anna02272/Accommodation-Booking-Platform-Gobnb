package utils

import (
	"bufio"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"os"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", fmt.Errorf("could not hash password %w", err)
	}
	return string(hashedPassword), nil
}

func VerifyPassword(hashedPassword string, candidatePassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(candidatePassword))
}
func VerifyNotHashedPassword(notHashedPassword string, candidatePassword string) error {
	if notHashedPassword != candidatePassword {
		return errors.New("passwords do not match")
	}
	return nil
}

func CheckBlackList(password string, filepath string) (bool, error) {
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Println(err.Error())
		return false, err
	}
	defer file.Close()

	blacklist := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		blacklist[line] = true
	}
	if blacklist[password] {
		return true, nil
	}
	return false, nil

}
