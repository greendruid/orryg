package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type client struct {
	endpoint string
	c        *http.Client
}

func newClient(endpoint string) (*client, error) {
	return &client{
		endpoint: endpoint,
		c:        &http.Client{},
	}, nil
}

func (c *client) post(path string, body interface{}) (string, error) {
	reqBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	resp, err := c.c.Post(c.endpoint+path, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unable to send request, got status %d. body was %s", resp.StatusCode, string(data))
	}

	return string(data), nil
}

func (c *client) postAndUnmarshal(path string, body interface{}, ptr interface{}) error {
	reqBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	resp, err := c.c.Post(c.endpoint+path, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unable to send request, got status %d. body was %s", resp.StatusCode, string(data))
	}

	return json.Unmarshal(data, ptr)
}
