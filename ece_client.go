package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const eceResource = "/api/v1/clusters/elasticsearch"
const jsonContentType = "application/json"

// ECEClient is a client used for interactions with the ECE API.
type ECEClient struct {
	// httpClient specifies the HTTP client that should be used for ECE API calls.
	httpClient *http.Client

	// url specifies the base URL for the ECE API.
	url string

	// username specifies the username to use for basic authentication.
	username string

	// password specifies the password to use for basic authentication.
	password string
}

// CreateCluster creates a new ECE deployment and cluster using the specified json string.
func (c *ECEClient) CreateCluster(json string) (resp *http.Response, err error) {
	body := strings.NewReader(json)

	log.Printf("[DEBUG] CreateCluster: %v\n", json)

	resourceURL := c.url + eceResource
	log.Printf("[DEBUG] CreateCluster Resource URL: %s\n", resourceURL)
	req, err := http.NewRequest("POST", resourceURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", jsonContentType)
	req.SetBasicAuth(c.username, c.password)

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] CreateCluster response: %v\n", resp)

	return resp, nil
}

// GetCluster returns the plan for an existing ECE deployment.
func (c *ECEClient) GetCluster(id string) (resp *http.Response, err error) {
	log.Printf("[DEBUG] GetCluster ID: %s\n", id)

	resourceURL := c.url + eceResource + "/" + id
	log.Printf("[DEBUG] GetCluster Resource URL: %s\n", resourceURL)
	req, err := http.NewRequest("GET", resourceURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", jsonContentType)
	req.SetBasicAuth(c.username, c.password)

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func getResponseAsJSON(resp *http.Response) (jsonResponse interface{}, err error) {
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bodyBytes, &jsonResponse)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] JSON Response: %v\n", jsonResponse)
	return jsonResponse, nil
}
