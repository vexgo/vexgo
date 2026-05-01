package auth

import (
	"testing"
)

func TestGenerateAndValidateToken(t *testing.T) {
	secret := "test-secret"
	SetSecret(secret)

	token, err := GenerateToken(1, "admin")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatal("expected token to be non-empty")
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("expected no error validating token, got %v", err)
	}
	if claims.UserID != 1 {
		t.Errorf("expected UserID 1, got %d", claims.UserID)
	}
	if claims.Role != "admin" {
		t.Errorf("expected Role admin, got %s", claims.Role)
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	SetSecret("test-secret")
	_, err := ValidateToken("invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}
