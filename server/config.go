package main

import "os"

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
	jwtSecret    = []byte(envOrDefault("JWT_SECRET", "dev-jwt-secret-change-me"))
)
