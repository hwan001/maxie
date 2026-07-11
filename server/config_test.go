package main

import (
	"strings"
	"testing"
)

// TestValidateSecrets verifies fail-fast on default/empty secrets in production
// while staying lenient in development.
func TestValidateSecrets(t *testing.T) {
	orig := jwtSecret
	t.Cleanup(func() { jwtSecret = orig })

	// Development: default secrets are tolerated.
	t.Setenv("APP_ENV", "")
	t.Setenv("GIN_MODE", "")
	jwtSecret = []byte(devJWTSecret)
	if err := validateSecrets(devSessionSecret); err != nil {
		t.Errorf("dev mode should tolerate default secrets, got %v", err)
	}

	// Production with default secrets: must fail and name both offenders.
	t.Setenv("APP_ENV", "production")
	jwtSecret = []byte(devJWTSecret)
	err := validateSecrets(devSessionSecret)
	if err == nil {
		t.Fatal("production with default secrets should fail")
	}
	if !strings.Contains(err.Error(), "JWT_SECRET") || !strings.Contains(err.Error(), "SESSION_SECRET") {
		t.Errorf("error should name both secrets, got %q", err.Error())
	}

	// Production with strong secrets: passes.
	jwtSecret = []byte("a-strong-unique-secret")
	if err := validateSecrets("another-strong-secret"); err != nil {
		t.Errorf("production with strong secrets should pass, got %v", err)
	}

	// Production with only the session secret left at default: fails, names it.
	jwtSecret = []byte("a-strong-unique-secret")
	err = validateSecrets(devSessionSecret)
	if err == nil || !strings.Contains(err.Error(), "SESSION_SECRET") {
		t.Errorf("expected failure naming SESSION_SECRET, got %v", err)
	}
}
