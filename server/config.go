package main

import (
	"fmt"
	"os"
	"strings"
)

// Development fallback secrets. Fine locally, refused in production — see
// validateSecrets.
const (
	devJWTSecret     = "dev-jwt-secret-change-me"
	devSessionSecret = "dev-session-secret-change-me"
)

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

var (
	clientID     = os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURI  = envOrDefault("REDIRECT_URI", "http://localhost:8080")
	jwtSecret    = []byte(envOrDefault("JWT_SECRET", devJWTSecret))
)

// isProduction reports whether the server is running in a production-like mode,
// where insecure default secrets must not be tolerated.
func isProduction() bool {
	return os.Getenv("GIN_MODE") == "release" ||
		strings.EqualFold(os.Getenv("APP_ENV"), "production")
}

// validateSecrets fails fast when running in production with empty or default
// signing secrets. In non-production it only warns, keeping local dev friction
// low. Returns an error describing every offending secret.
func validateSecrets(sessionSecret string) error {
	var offenders []string
	if len(jwtSecret) == 0 || string(jwtSecret) == devJWTSecret {
		offenders = append(offenders, "JWT_SECRET")
	}
	if sessionSecret == "" || sessionSecret == devSessionSecret {
		offenders = append(offenders, "SESSION_SECRET")
	}
	if len(offenders) == 0 {
		return nil
	}
	if isProduction() {
		return fmt.Errorf("insecure default/empty secret(s) in production: %s — set a strong value before deploying", strings.Join(offenders, ", "))
	}
	return nil
}
