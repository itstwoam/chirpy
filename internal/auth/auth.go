package auth

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"time"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

func HashPassword(password string) (string, error) {
	hashWord, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
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

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn int) (string, error) {
	dExpiresIn := time.Duration(expiresIn) * time.Second
	newClaims := jwt.RegisteredClaims{
		Issuer: "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(dExpiresIn)),
		Subject:   userID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	userToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", errors.New("error retrieving user token")
	}
	return userToken, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	bToken := headers["Authorization"]
	if bToken == nil {
		return "", errors.New("Invalid/missing bearer token")
	}
	sstrings := strings.Split(bToken[0], " ")
	return strings.TrimSpace(sstrings[1]), nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		fmt.Printf("Unexpected signing method: %v", token.Header["alg"])
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return []byte(tokenSecret), nil
})
	if err != nil {
		return uuid.Nil, errors.New("failed to validate token")
	}
	if !token.Valid {
		return uuid.Nil, errors.New("invalid token")
	}
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return uuid.Nil, errors.New("invalid claim")
	}
	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, errors.New("invalid id")
	}
	return id, nil
}
