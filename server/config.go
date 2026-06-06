package main

import (
	"log"
	"os"
)

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("%s environment variable is required", key)
	}
	return v
}

var (
	clientID     = os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURI  = envOrDefault("REDIRECT_URI", "http://localhost:8080")
	jwtSecret    = []byte(requireEnv("JWT_SECRET"))
)
