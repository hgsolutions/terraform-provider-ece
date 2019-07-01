package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
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

	// timeout in seconds for resource operations.
	timeout int
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

// DeleteCluster deletes an existing ECE deployment and cluster using the specified json string.
func (c *ECEClient) DeleteCluster(id string) (resp *http.Response, err error) {
	log.Printf("[DEBUG] DeleteCluster ID: %s\n", id)

	resourceURL := c.url + eceResource
	log.Printf("[DEBUG] DeleteCluster Resource URL: %s\n", resourceURL)
	req, err := http.NewRequest("DELETE", resourceURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", jsonContentType)
	req.SetBasicAuth(c.username, c.password)

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] DeleteCluster response: %v\n", resp)

	return resp, nil
}

// GetCluster returns the metadata for an existing ECE deployment.
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

	log.Printf("[DEBUG] GetCluster response: %v\n", resp)

	return resp, nil
}

// GetResponseAsJSON returns a response body as a JSON document.
func (c *ECEClient) GetResponseAsJSON(resp *http.Response) (jsonResponse interface{}, err error) {
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bodyBytes, &jsonResponse)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Error unmarshalling response body: %+v: %+v", err, resp.Body)
	}

	log.Printf("[DEBUG] JSON Response: %v\n", jsonResponse)
	return jsonResponse, nil
}

// ShutdownCluster shuts down an existing ECE deployment and cluster.
func (c *ECEClient) ShutdownCluster(id string) (resp *http.Response, err error) {
	log.Printf("[DEBUG] ShutdownCluster ID: %s\n", id)

	resourceURL := c.url + eceResource + "/_shutdown"
	log.Printf("[DEBUG] Shutdown resource URL: %s\n", resourceURL)
	req, err := http.NewRequest("POST", resourceURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", jsonContentType)
	req.SetBasicAuth(c.username, c.password)

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] ShutdownCluster response: %v\n", resp)

	return resp, nil
}

// WaitForCreate waits for a cluster to be created.
func (c *ECEClient) WaitForCreate(id string) error {
	timeoutSeconds := time.Second * time.Duration(c.timeout)
	log.Printf("[DEBUG] WaitForCreate will wait for %v seconds for creation of cluster ID: %s\n", timeoutSeconds, id)

	return resource.Retry(timeoutSeconds, func() *resource.RetryError {
		resp, err := c.GetCluster(id)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		if resp.StatusCode == 200 {
			respBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return resource.NonRetryableError(err)
			}

			var clusterInfo ElasticsearchClusterInfo
			err = json.Unmarshal(respBytes, &clusterInfo)
			if err != nil {
				return resource.NonRetryableError(err)
			}

			log.Printf("[DEBUG] WaitForCreate cluster status: %s\n", clusterInfo.Status)

			if clusterInfo.Status == "started" {
				return nil
			}
		}

		return resource.RetryableError(
			fmt.Errorf("%q: Timeout while waiting for the cluster to be created", id))
	})
}

// WaitForShutdown waits for a cluster to shutdown.
func (c *ECEClient) WaitForShutdown(id string) error {
	timeoutSeconds := time.Second * time.Duration(c.timeout)
	log.Printf("[DEBUG] WaitForShutdown will wait for %v seconds for shutdown of cluster ID: %s\n", timeoutSeconds, id)

	return resource.Retry(timeoutSeconds, func() *resource.RetryError {
		resp, err := c.GetCluster(id)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		if resp.StatusCode == 404 {
			return nil
		} else if resp.StatusCode == 200 {
			respBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return resource.NonRetryableError(err)
			}

			var clusterInfo ElasticsearchClusterInfo
			err = json.Unmarshal(respBytes, &clusterInfo)
			if err != nil {
				return resource.NonRetryableError(err)
			}

			log.Printf("[DEBUG] WaitForShutdown cluster status: %s\n", clusterInfo.Status)

			if clusterInfo.Status == "stopped" {
				return nil
			}
		}

		return resource.RetryableError(
			fmt.Errorf("%q: Timeout while waiting for the cluster to shutdown", id))
	})
}
