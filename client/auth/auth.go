package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"
	"fileoptimizer/common"
)

// Authenticate sends a login request to the server and retrieves the JWT token
func Authenticate(serverURL string, creds common.Credentials) (string, error) {
	url := serverURL + "/auth/login"
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	credsJson, err := json.Marshal(creds)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(credsJson))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to authenticate with server")
	}

	var authResp common.AuthResponse
	err = json.NewDecoder(resp.Body).Decode(&authResp)
	if err != nil {
		return "", err
	}

	return authResp.Token, nil
}
