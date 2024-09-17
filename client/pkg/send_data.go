package pkg

import (
	"errors"
	"bytes"
	"net/http"
	"time"
)

func SendData(serverURL string, encryptedData string, token string) error {
	url := serverURL + "/data/upload"
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(encryptedData)))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to send data to the server")
	}

	return nil
}
