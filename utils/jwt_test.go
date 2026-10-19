package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/personal-stakshi/draftly/backend/config"
)

func testConfig() *config.Config {
	return &config.Config{JWTSecret: "test-secret"}
}

func TestGenerateTokenAndValidate(t *testing.T) {
	userID := uuid.New()
	cfg := testConfig()

	token, err := GenerateToken(userID, cfg)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("expected a token, got empty string")
	}

	claims, err := ValidateToken(token, cfg)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.UserID != userID {
		t.Fatalf("user id mismatch: got %s want %s", claims.UserID, userID)
	}
	if claims.Issuer != "draftly-backend" {
		t.Fatalf("issuer mismatch: got %s", claims.Issuer)
	}
}

func TestValidateTokenWrongSecret(t *testing.T) {
	token, err := GenerateToken(uuid.New(), testConfig())
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if _, err := ValidateToken(token, &config.Config{JWTSecret: "other-secret"}); err == nil {
		t.Fatal("expected error for a token signed with a different secret")
	}
}

func TestValidateTokenExpired(t *testing.T) {
	cfg := testConfig()
	claims := &Claims{
		UserID: uuid.New(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		t.Fatalf("signing failed: %v", err)
	}

	if _, err := ValidateToken(token, cfg); err == nil {
		t.Fatal("expected error for an expired token")
	}
}

func TestValidateTokenGarbage(t *testing.T) {
	if _, err := ValidateToken("not-a-real-token", testConfig()); err == nil {
		t.Fatal("expected error for a malformed token")
	}
}

func TestValidateTokenRejectsWrongSigningMethod(t *testing.T) {
	claims := &Claims{UserID: uuid.New()}
	token, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("signing failed: %v", err)
	}

	if _, err := ValidateToken(token, testConfig()); err == nil {
		t.Fatal("expected error for a token that is not signed with HMAC")
	}
}
