package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"
	expires := time.Hour

	tokenStr, err := MakeJWT(userID, secret, expires)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	// Parse and validate the token
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		t.Fatalf("Token is invalid: %v", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		t.Fatal("Failed to parse claims")
	}

	if claims.Subject != userID.String() {
		t.Errorf("Expected subject %s, got %s", userID.String(), claims.Subject)
	}
}
