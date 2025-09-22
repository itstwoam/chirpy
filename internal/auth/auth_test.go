package auth

import (
	"fmt"
	"testing"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"
	expires := 3600

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

func TestMakeRefreshToken(t *testing.T) {
	key := MakeRefreshToken()
	fmt.Println(key)
	if len(key) != 64 {
		t.Fatalf("Key length is %v, when it should be 32", len(key))
	}
}

