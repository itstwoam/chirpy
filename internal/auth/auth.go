package auth

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"time"
	//"errors"
)

func HashPassword(password string) (string, error) {
	start := time.Now()
	hashWord, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	fmt.Println("Hash time:", time.Since(start))
	if err != nil {
		fmt.Println("error while hashing password")
		return "unset", err
	}
	return string(hashWord), nil
}

func CheckPasswordHash(password, hash string) error {
	modPassword := []byte(password)
	modHash := []byte(hash)
	return bcrypt.CompareHashAndPassword(modHash, modPassword)
}
