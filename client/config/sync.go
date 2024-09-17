package config

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func SyncConfig(serverURL, token string) error {
	url := serverURL + "/config/sync"
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to sync config from server")
	}

	// Read and write the new config to the local file
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("config.json", body, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
