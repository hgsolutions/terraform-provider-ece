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

// CreateCluster creates a new ECE cluster using the specified create request.
func (c *ECEClient) CreateCluster(createClusterRequest CreateElasticsearchClusterRequest) (crudResponse *ClusterCrudResponse, err error) {
	log.Printf("[DEBUG] CreateCluster: %v\n", createClusterRequest)

	// Example cluster creation request body.
	// {
	// 	"cluster_name" : "My Cluster",
	// 	"plan" : {
	// 		"elasticsearch" : {
	// 			"version" : "7.1.0"
	// 		},
	// 		"cluster_topology" : [
	// 			{
	// 				"memory_per_node" : 2048,
	// 				"node_count_per_zone" : 1,
	// 				"node_type" : {
	// 				   "data" : true,
	// 				   "ingest" : true,
	// 				   "master" : true,
	// 				   "ml" : true
	// 				},
	// 				"zone_count" : 1
	// 			}
	// 		]
	// 	 }
	// }

	jsonData, err := json.Marshal(createClusterRequest)
	if err != nil {
		return nil, err
	}

	jsonString := string(jsonData)
	body := strings.NewReader(jsonString)

	resourceURL := c.url + eceResource
	log.Printf("[DEBUG] CreateCluster Resource URL: %s\n", resourceURL)
	req, err := http.NewRequest("POST", resourceURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", jsonContentType)
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Example response:
	// {
	// 	"elasticsearch_cluster_id": "5de00f3876e3442f8e4f83110af0e251",
	// 	"credentials": {
	// 		"username": "elastic",
	// 		"password": "Ov8cmAVCqTr8biFfND2wtIuY"
	// 	}
	// }

	log.Printf("[DEBUG] CreateCluster response: %v\n", resp)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("cluster could not be created: %v", string(respBytes))
	}

	log.Printf("[DEBUG] CreateCluster response body: %v\n", string(respBytes))

	err = json.Unmarshal(respBytes, &crudResponse)
	if err != nil {
		return nil, err
	}

	return crudResponse, nil
}

// DeleteCluster deletes an existing ECE cluster.
func (c *ECEClient) DeleteCluster(id string) (resp *http.Response, err error) {
	log.Printf("[DEBUG] DeleteCluster ID: %s\n", id)

	resourceURL := c.url + eceResource + "/" + id
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

	if resp.StatusCode != 200 {
		respBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%q: cluster could not be deleted: %v", id, string(respBytes))
	}

	return resp, nil
}

// GetCluster returns information for an existing ECE cluster.
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

	if resp.StatusCode != 200 && resp.StatusCode != 404 {
		respBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%q: cluster could not be retrieved: %v", id, string(respBytes))
	}

	return resp, nil
}

// GetResponseBodyAsJSON returns a response body as a JSON document.
func (c *ECEClient) GetResponseBodyAsJSON(resp *http.Response) (jsonResponse interface{}, err error) {
	err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %+v: %+v", err, resp.Body)
	}

	log.Printf("[DEBUG] GetResponseBodyAsJSON body: %v\n", jsonResponse)

	return jsonResponse, nil
}

// GetResponseBodyAsString returns a response body as a string.
func (c *ECEClient) GetResponseBodyAsString(resp *http.Response) (body string, err error) {
	log.Printf("[DEBUG] GetResponseBodyAsString: %v\n", resp)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	body = string(respBytes)

	log.Printf("[DEBUG] GetResponseBodyAsString body: %v\n", body)

	return body, nil
}

// UpdateCluster updates an existing ECE cluster using the specified cluster plan.
func (c *ECEClient) UpdateCluster(id string, clusterPlan ElasticsearchClusterPlan) (resp *http.Response, err error) {
	log.Printf("[DEBUG] UpdateCluster: %v\n", clusterPlan)

	jsonData, err := json.Marshal(clusterPlan)
	if err != nil {
		return nil, err
	}

	jsonString := string(jsonData)
	body := strings.NewReader(jsonString)

	resourceURL := c.url + eceResource + "/" + id + "/plan"
	log.Printf("[DEBUG] UpdateCluster Resource URL: %s\n", resourceURL)
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

	log.Printf("[DEBUG] UpdateCluster response: %v\n", resp)

	if resp.StatusCode != 202 {
		respBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%q: cluster could not be updated: %v", id, string(respBytes))
	}

	return resp, nil
}

// ShutdownCluster shuts down an existing ECE cluster.
func (c *ECEClient) ShutdownCluster(id string) (resp *http.Response, err error) {
	log.Printf("[DEBUG] ShutdownCluster ID: %s\n", id)

	resourceURL := c.url + eceResource + "/" + id + "/_shutdown"
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

	if resp.StatusCode != 202 {
		respBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%q: cluster could not be shutdown: %v", id, string(respBytes))
	}

	return resp, nil
}

// WaitForStatus waits for a cluster to enter the specified status.
func (c *ECEClient) WaitForStatus(id string, status string) error {
	timeoutSeconds := time.Second * time.Duration(c.timeout)
	log.Printf("[DEBUG] WaitForStatus will wait for %v seconds for '%s' status for cluster ID: %s\n", timeoutSeconds, status, id)

	return resource.Retry(timeoutSeconds, func() *resource.RetryError {
		resp, err := c.GetCluster(id)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		if resp.StatusCode == 200 {
			var clusterInfo ElasticsearchClusterInfo
			err = json.NewDecoder(resp.Body).Decode(&clusterInfo)
			if err != nil {
				return resource.NonRetryableError(err)
			}

			if clusterInfo.Status == status {
				log.Printf("[DEBUG] WaitForStatus desired cluster status reached: %s\n", clusterInfo.Status)
				return nil
			}

			log.Printf("[DEBUG] WaitForStatus current cluster status: %s. Desired status: %s\n", clusterInfo.Status, status)
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
			var clusterInfo ElasticsearchClusterInfo
			err = json.NewDecoder(resp.Body).Decode(&clusterInfo)
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
