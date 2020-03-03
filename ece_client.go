package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
)

const baseEndpoint = "/api/v1"
const elasticsearchResource = baseEndpoint + "/clusters/elasticsearch"
const kibanaResource = baseEndpoint + "/clusters/kibana"
const deploymentResource = baseEndpoint + "/deployments"
const jsonContentType = "application/json"

// ECEClient is a client used for interactions with the ECE API.
type ECEClient struct {
	// HTTPClient specifies the HTTP client that should be used for ECE API calls.
	HTTPClient *http.Client

	// BaseURL specifies the base URL for the ECE API.
	BaseURL string

	// Username specifies the Username to use for basic authentication.
	Username string

	// Password specifies the Password to use for basic authentication.
	Password string

	// Timeout in seconds for resource operations.
	Timeout int

	// Token to be used for requests as a bearer token.
	AuthToken string

	// True if interacting wtih Elastic-Cloud.
	IsElasticCloud bool
}

// BearerToken constructs and returns the authentication header.
func (c *ECEClient) BearerToken() string {
	return "Bearer " + c.AuthToken
}

// SetRequestAuth Conditionally sets the header for ECE or Elastic-Cloud.
func (c *ECEClient) SetRequestAuth(req *http.Request) {
	if c.IsElasticCloud {
		req.Header.Set("Authorization", c.BearerToken())
	} else {
		req.SetBasicAuth(c.Username, c.Password)
	}
}

// Login attempts to log in using username and password sets the ECEClient's AuthToken.
func (c *ECEClient) Login() (err error) {
	log.Printf("[DEBUG] LoginToECE : %s\n", c.Username)

	// Note: This URL is different from the documented /api/v1/users/auth/_login.
	resourceURL := c.BaseURL + baseEndpoint + "/users/_login"
	log.Printf("[DEBUG] LoginToECE request url: %s\n", resourceURL)

	loginRequest := LoginRequest{
		Password: c.Password,
		Username: c.Username,
	}

	jsonData, err := json.Marshal(loginRequest)
	if err != nil {
		return err
	}

	jsonString := string(jsonData)

	body := strings.NewReader(jsonString)
	req, err := http.NewRequest("POST", resourceURL, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", jsonContentType)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] LoginToECE response: %v\n", resp)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("%q: LoginToECE failed: %v", c.Username, string(respBytes))
	}

	var token TokenResponse
	err = json.Unmarshal(respBytes, &token)
	if err != nil {
		return err
	}

	c.AuthToken = token.Token

	return nil
}

// CreateDeployment creates a new deployment using the specified create request.
func (c *ECEClient) CreateDeployment(deploymentCreateRequest DeploymentCreateRequest) (deploymentCreateResponse *DeploymentCreateResponse, err error) {
	jsonData, err := json.Marshal(deploymentCreateRequest)
	if err != nil {
		return nil, err
	}

	jsonString := string(jsonData)
	log.Printf("[DEBUG] CreateDeployment request body: %s\n", jsonString)

	body := strings.NewReader(jsonString)
	resourceURL := c.BaseURL + deploymentResource
	log.Printf("[DEBUG] CreateDeployment Resource URL: %s\n", resourceURL)
	req, err := http.NewRequest("POST", resourceURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", jsonContentType)
	c.SetRequestAuth(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] CreateDeployment response: %v\n", resp)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("CreateDeployment failed: %v", string(respBytes))
	}

	log.Printf("[DEBUG] CreateDeployment response body: %v\n", string(respBytes))

	err = json.Unmarshal(respBytes, &deploymentCreateResponse)
	if err != nil {
		return nil, err
	}

	return deploymentCreateResponse, err
}

// GetDeployment returns information for an existing deployment.
func (c *ECEClient) GetDeployment(id string) (resp *http.Response, err error) {
	log.Printf("[DEBUG] GetDeployment ID: %s\n", id)

	resourceURL := c.BaseURL + deploymentResource + "/" + id
	log.Printf("[DEBUG] GetDeployment Resource URL: %s\n", resourceURL)
	req, err := http.NewRequest("GET", resourceURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", jsonContentType)
	c.SetRequestAuth(req)

	resp, err = c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] GetDeployment response: %v\n", resp)

	if resp.StatusCode != 200 && resp.StatusCode != 404 {
		respBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%q: deployment could not be retrieved: %v", id, string(respBytes))
	}

	return resp, nil
}

// DeleteDeployment deletes an existing deployment.
func (c *ECEClient) DeleteDeployment(id string) (resp *http.Response, err error) {
	log.Printf("[DEBUG] DeleteDeployment ID: %s\n", id)

	// NOTE: A deployment must be successfully _shutdown first before it can be deleted.
	log.Printf("[DEBUG] Deleting deployment ID: %s\n", id)
	resp, err = c.ShutdownDeployment(id, true, true)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// ShutdownDeployment shuts down an existing deployment.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/Deployment_-_CRUD.html#shutdown-deployment
func (c *ECEClient) ShutdownDeployment(id string, hide bool, skipSnapshot bool) (resp *http.Response, err error) {
	log.Printf("[DEBUG] ShutdownDeployment ID: %s\n", id)

	resourceURL := c.BaseURL + deploymentResource + "/" + id + "/_shutdown?hide=" + strconv.FormatBool(hide) + "&skip_snapshot=" + strconv.FormatBool(skipSnapshot)
	log.Printf("[DEBUG] ShutdownDeployment resource URL: %s\n", resourceURL)
	req, err := http.NewRequest("POST", resourceURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", jsonContentType)
	c.SetRequestAuth(req)

	resp, err = c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] ShutdownDeployment response: %v\n", resp)

	if resp.StatusCode != 202 {
		respBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%q: elasticsearch cluster could not be shutdown: %v", id, string(respBytes))
	}

	return resp, nil
}

// WaitForDeploymentStatus waits for a deployment to be deleted.
func (c *ECEClient) WaitForDeploymentStatus(id string, allowMissing bool) error {
	timeoutSeconds := time.Second * time.Duration(c.Timeout)
	log.Printf("[DEBUG] WaitForDeploymentStatus will wait for %v seconds for deployment ID: %s\n", timeoutSeconds, id)

	return resource.Retry(timeoutSeconds, func() *resource.RetryError {
		resp, err := c.GetDeployment(id)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		if resp.StatusCode == 404 && allowMissing {
			return nil
		}

		return resource.RetryableError(
			fmt.Errorf("%q: timeout while waiting for the deployment to shutdown", id))
	})
}

// CreateElasticsearchCluster creates a new elasticsearch cluster using the specified create request.
func (c *ECEClient) CreateElasticsearchCluster(createClusterRequest CreateElasticsearchClusterRequest) (crudResponse *ClusterCrudResponse, err error) {
	log.Printf("[DEBUG] CreateElasticsearchCluster: %v\n", createClusterRequest)

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
	log.Printf("[DEBUG] CreateElasticsearchCluster request body: %s\n", jsonString)

	body := strings.NewReader(jsonString)
	resourceURL := c.BaseURL + elasticsearchResource
	log.Printf("[DEBUG] CreateElasticsearchCluster Resource URL: %s\n", resourceURL)
	req, err := http.NewRequest("POST", resourceURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", jsonContentType)
	c.SetRequestAuth(req)

	resp, err := c.HTTPClient.Do(req)
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

	log.Printf("[DEBUG] CreateElasticsearchCluster response: %v\n", resp)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("elasticsearch cluster could not be created: %v", string(respBytes))
	}

	log.Printf("[DEBUG] CreateElasticsearchCluster response body: %v\n", string(respBytes))

	err = json.Unmarshal(respBytes, &crudResponse)
	if err != nil {
		return nil, err
	}

	return crudResponse, nil
}

// CreateKibanaCluster creates a new Kibana cluster using the specified create request.
func (c *ECEClient) CreateKibanaCluster(createKibanaRequest CreateKibanaRequest) (crudResponse *ClusterCrudResponse, err error) {
	log.Printf("[DEBUG] CreateKibanaCluster: %v\n", createKibanaRequest)

	jsonData, err := json.Marshal(createKibanaRequest)
	if err != nil {
		return nil, err
	}

	jsonString := string(jsonData)
	log.Printf("[DEBUG] CreateKibanaCluster request body: %s\n", jsonString)

	body := strings.NewReader(jsonString)
	resourceURL := c.BaseURL + kibanaResource
	log.Printf("[DEBUG] CreateKibanaCluster Resource URL: %s\n", resourceURL)
	req, err := http.NewRequest("POST", resourceURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", jsonContentType)
	c.SetRequestAuth(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] CreateKibanaCluster response: %v\n", resp)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("kibana cluster could not be created: %v", string(respBytes))
	}

	log.Printf("[DEBUG] CreateKibanaCluster response body: %v\n", string(respBytes))

	err = json.Unmarshal(respBytes, &crudResponse)
	if err != nil {
		return nil, err
	}

	return crudResponse, nil
}

// DeleteElasticsearchCluster deletes an existing elasticsearch cluster.
func (c *ECEClient) DeleteElasticsearchCluster(id string) (resp *http.Response, err error) {
	log.Printf("[DEBUG] DeleteElasticsearchCluster ID: %s\n", id)

	// NOTE: A cluster must be successfully _shutdown first before it can be deleted.
	log.Printf("[DEBUG] Shutting down cluster ID: %s\n", id)
	_, err = c.ShutdownElasticsearchCluster(id)
	if err != nil {
		return nil, err
	}

	// Wait for cluster shutdown.
	log.Printf("[DEBUG] Waiting for shutdown of cluster ID: %s\n", id)
	c.WaitForElasticsearchClusterStatus(id, "stopped", true)

	resourceURL := c.BaseURL + elasticsearchResource + "/" + id
	log.Printf("[DEBUG] DeleteElasticsearchCluster Resource URL: %s\n", resourceURL)
	req, err := http.NewRequest("DELETE", resourceURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", jsonContentType)
	c.SetRequestAuth(req)

	resp, err = c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] DeleteElasticsearchCluster response: %v\n", resp)

	if resp.StatusCode != 200 {
		respBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%q: elasticsearch cluster could not be deleted: %v", id, string(respBytes))
	}

	return resp, nil
}

// DeleteKibanaCluster deletes an existing kibana cluster.
func (c *ECEClient) DeleteKibanaCluster(id string) (resp *http.Response, err error) {
	log.Printf("[DEBUG] DeleteKibanaCluster ID: %s\n", id)

	// NOTE: A cluster must be successfully _shutdown first before it can be deleted.
	log.Printf("[DEBUG] Shutting down cluster ID: %s\n", id)
	_, err = c.ShutdownKibanaCluster(id)
	if err != nil {
		return nil, err
	}

	// Wait for cluster shutdown.
	log.Printf("[DEBUG] Waiting for shutdown of cluster ID: %s\n", id)
	c.WaitForKibanaClusterStatus(id, "stopped", true)

	resourceURL := c.BaseURL + kibanaResource + "/" + id
	log.Printf("[DEBUG] DeleteKibanaCluster Resource URL: %s\n", resourceURL)
	req, err := http.NewRequest("DELETE", resourceURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", jsonContentType)
	c.SetRequestAuth(req)

	resp, err = c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] DeleteKibanaCluster response: %v\n", resp)

	if resp.StatusCode != 200 {
		respBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%q: kibana cluster could not be deleted: %v", id, string(respBytes))
	}

	return resp, nil
}

// GetElasticsearchCluster returns information for an existing elasticsearch cluster.
func (c *ECEClient) GetElasticsearchCluster(id string) (resp *http.Response, err error) {
	log.Printf("[DEBUG] GetElasticsearchCluster ID: %s\n", id)

	resourceURL := c.BaseURL + elasticsearchResource + "/" + id
	log.Printf("[DEBUG] GetElasticsearchCluster Resource URL: %s\n", resourceURL)
	req, err := http.NewRequest("GET", resourceURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", jsonContentType)
	c.SetRequestAuth(req)

	resp, err = c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] GetElasticsearchCluster response: %v\n", resp)

	if resp.StatusCode != 200 && resp.StatusCode != 404 {
		respBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%q: elasticsearch cluster could not be retrieved: %v", id, string(respBytes))
	}

	return resp, nil
}

// GetElasticsearchClusterPlan returns the plan for an existing elasticsearch cluster.
func (c *ECEClient) GetElasticsearchClusterPlan(id string) (resp *http.Response, err error) {
	log.Printf("[DEBUG] GetElasticsearchClusterPlan ID: %s\n", id)

	// GET /api/v1/clusters/elasticsearch/{cluster_id}/plan
	resourceURL := c.BaseURL + elasticsearchResource + "/" + id + "/plan"
	log.Printf("[DEBUG] GetElasticsearchClusterPlan Resource URL: %s\n", resourceURL)
	req, err := http.NewRequest("GET", resourceURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", jsonContentType)
	c.SetRequestAuth(req)

	resp, err = c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] GetElasticsearchClusterPlan response: %v\n", resp)

	if resp.StatusCode != 200 && resp.StatusCode != 404 {
		respBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%q: elasticsearch cluster plan could not be retrieved: %v", id, string(respBytes))
	}

	return resp, nil
}

// GetElasticsearchClusterPlanActivity returns the active and historical plan information for an elasticsearch cluster.
func (c *ECEClient) GetElasticsearchClusterPlanActivity(id string) (resp *http.Response, err error) {
	log.Printf("[DEBUG] GetElasticsearchClusterPlanActivity ID: %s\n", id)

	// GET /api/v1/clusters/elasticsearch/{cluster_id}/plan/activity
	resourceURL := c.BaseURL + elasticsearchResource + "/" + id + "/plan/activity"
	log.Printf("[DEBUG] GetElasticsearchClusterPlanActivity Resource URL: %s\n", resourceURL)
	req, err := http.NewRequest("GET", resourceURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", jsonContentType)
	c.SetRequestAuth(req)

	resp, err = c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] GetElasticsearchClusterPlanActivity response: %v\n", resp)

	if resp.StatusCode != 200 && resp.StatusCode != 404 {
		respBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%q: elasticsearch cluster plan activity could not be retrieved: %v", id, string(respBytes))
	}

	return resp, nil
}

// GetKibanaCluster returns information for an existing Kibana cluster.
func (c *ECEClient) GetKibanaCluster(id string) (resp *http.Response, err error) {
	log.Printf("[DEBUG] GetKibanaCluster ID: %s\n", id)

	resourceURL := c.BaseURL + kibanaResource + "/" + id
	log.Printf("[DEBUG] GetKibanaCluster Resource URL: %s\n", resourceURL)
	req, err := http.NewRequest("GET", resourceURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", jsonContentType)
	c.SetRequestAuth(req)

	resp, err = c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] GetKibanaCluster response: %v\n", resp)

	if resp.StatusCode != 200 && resp.StatusCode != 404 {
		respBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%q: kibana cluster could not be retrieved: %v", id, string(respBytes))
	}

	return resp, nil
}

// GetKibanaClusterPlanActivity returns the active and historical plan information for a Kibana cluster.
func (c *ECEClient) GetKibanaClusterPlanActivity(id string) (resp *http.Response, err error) {
	log.Printf("[DEBUG] GetKibanaClusterPlanActivity ID: %s\n", id)

	// GET /api/v1/clusters/kibana/{cluster_id}/plan/activity
	resourceURL := c.BaseURL + kibanaResource + "/" + id + "/plan/activity"
	log.Printf("[DEBUG] GetKibanaClusterPlanActivity Resource URL: %s\n", resourceURL)
	req, err := http.NewRequest("GET", resourceURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", jsonContentType)
	c.SetRequestAuth(req)

	resp, err = c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] GetKibanaClusterPlanActivity response: %v\n", resp)

	if resp.StatusCode != 200 && resp.StatusCode != 404 {
		respBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%q: kibana cluster plan activity could not be retrieved: %v", id, string(respBytes))
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

// UpdateElasticsearchCluster updates an existing elasticsearch cluster using the specified cluster plan.
func (c *ECEClient) UpdateElasticsearchCluster(id string, clusterPlan ElasticsearchClusterPlan) (resp *http.Response, err error) {
	log.Printf("[DEBUG] UpdateElasticsearchCluster: %s: %v\n", id, clusterPlan)

	jsonData, err := json.Marshal(clusterPlan)
	if err != nil {
		return nil, err
	}

	jsonString := string(jsonData)
	body := strings.NewReader(jsonString)

	resourceURL := c.BaseURL + elasticsearchResource + "/" + id + "/plan"
	log.Printf("[DEBUG] UpdateElasticsearchCluster Resource URL: %s\n", resourceURL)
	req, err := http.NewRequest("POST", resourceURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", jsonContentType)
	c.SetRequestAuth(req)

	resp, err = c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] UpdateElasticsearchCluster response: %v\n", resp)

	if resp.StatusCode != 202 {
		respBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%q: elasticsearch cluster could not be updated: %v", id, string(respBytes))
	}

	return resp, nil
}

// UpdateElasticsearchClusterMetadata updates the metadata for an existing elasticsearch cluster.
func (c *ECEClient) UpdateElasticsearchClusterMetadata(id string, metadata ClusterMetadataSettings) (resp *http.Response, err error) {
	log.Printf("[DEBUG] UpdateElasticsearchClusterMetadata: %s: %v\n", id, metadata)

	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	jsonString := string(jsonData)
	body := strings.NewReader(jsonString)

	// PATCH /api/v1/clusters/elasticsearch/{cluster_id}/metadata/settings
	resourceURL := c.BaseURL + elasticsearchResource + "/" + id + "/metadata/settings"
	log.Printf("[DEBUG] UpdateElasticsearchClusterMetadata Resource URL: %s\n", resourceURL)
	req, err := http.NewRequest("PATCH", resourceURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", jsonContentType)
	c.SetRequestAuth(req)

	resp, err = c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] UpdateElasticsearchClusterMetadata response: %v\n", resp)

	if resp.StatusCode != 200 {
		respBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%q: elasticsearch cluster metadata settings could not be updated: %v", id, string(respBytes))
	}

	return resp, nil
}

// UpdateKibanaCluster updates an existing Kibana cluster using the specified Kibana cluster plan.
func (c *ECEClient) UpdateKibanaCluster(id string, kibanaPlan *KibanaClusterPlan) (resp *http.Response, err error) {
	log.Printf("[DEBUG] UpdateKibanaCluster: %s: %v\n", id, *kibanaPlan)

	jsonData, err := json.Marshal(kibanaPlan)
	if err != nil {
		return nil, err
	}

	jsonString := string(jsonData)
	body := strings.NewReader(jsonString)

	resourceURL := c.BaseURL + kibanaResource + "/" + id + "/plan"
	log.Printf("[DEBUG] UpdateKibanaCluster Resource URL: %s\n", resourceURL)
	req, err := http.NewRequest("POST", resourceURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", jsonContentType)
	c.SetRequestAuth(req)

	resp, err = c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] UpdateKibanaCluster response: %v\n", resp)

	if resp.StatusCode != 202 {
		respBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%q: kibana cluster could not be updated: %v", id, string(respBytes))
	}

	return resp, nil
}

// UpdateKibanaClusterMetadata updates the metadata for an existing Kibana cluster.
func (c *ECEClient) UpdateKibanaClusterMetadata(id string, metadata ClusterMetadataSettings) (resp *http.Response, err error) {
	log.Printf("[DEBUG] UpdateKibanaClusterMetadata: %s: %v\n", id, metadata)

	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	jsonString := string(jsonData)
	body := strings.NewReader(jsonString)

	// PATCH /api/v1/clusters/kibana/{cluster_id}/metadata/settings
	resourceURL := c.BaseURL + kibanaResource + "/" + id + "/metadata/settings"
	log.Printf("[DEBUG] UpdateKibanaClusterMetadata Resource URL: %s\n", resourceURL)
	req, err := http.NewRequest("PATCH", resourceURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", jsonContentType)
	c.SetRequestAuth(req)

	resp, err = c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] UpdateKibanaClusterMetadata response: %v\n", resp)

	if resp.StatusCode != 200 {
		respBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%q: kibana cluster metadata settings could not be updated: %v", id, string(respBytes))
	}

	return resp, nil
}

// ShutdownElasticsearchCluster shuts down an existing ECE cluster.
func (c *ECEClient) ShutdownElasticsearchCluster(id string) (resp *http.Response, err error) {
	log.Printf("[DEBUG] ShutdownElasticsearchCluster ID: %s\n", id)

	resourceURL := c.BaseURL + elasticsearchResource + "/" + id + "/_shutdown"
	log.Printf("[DEBUG] ShutdownElasticsearchCluster resource URL: %s\n", resourceURL)
	req, err := http.NewRequest("POST", resourceURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", jsonContentType)
	c.SetRequestAuth(req)

	resp, err = c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] ShutdownElasticsearchCluster response: %v\n", resp)

	if resp.StatusCode != 202 {
		respBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%q: elasticsearch cluster could not be shutdown: %v", id, string(respBytes))
	}

	return resp, nil
}

// ShutdownKibanaCluster shuts down an existing Kibana cluster.
func (c *ECEClient) ShutdownKibanaCluster(id string) (resp *http.Response, err error) {
	log.Printf("[DEBUG] ShutdownKibanaCluster ID: %s\n", id)

	resourceURL := c.BaseURL + kibanaResource + "/" + id + "/_shutdown"
	log.Printf("[DEBUG] ShutdownKibanaCluster resource URL: %s\n", resourceURL)
	req, err := http.NewRequest("POST", resourceURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", jsonContentType)
	c.SetRequestAuth(req)

	resp, err = c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] ShutdownKibanaCluster response: %v\n", resp)

	if resp.StatusCode != 202 {
		respBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%q: kibana cluster could not be shutdown: %v", id, string(respBytes))
	}

	return resp, nil
}

// WaitForElasticsearchClusterStatus waits for an elasticsearch cluster to enter the specified status.
func (c *ECEClient) WaitForElasticsearchClusterStatus(id string, status string, allowMissing bool) error {
	timeoutSeconds := time.Second * time.Duration(c.Timeout)
	log.Printf("[DEBUG] WaitForElasticsearchClusterStatus will wait for %v seconds for '%s' status for cluster ID: %s\n", timeoutSeconds, status, id)

	return resource.Retry(timeoutSeconds, func() *resource.RetryError {
		resp, err := c.GetElasticsearchCluster(id)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		if resp.StatusCode == 404 && allowMissing {
			return nil
		} else if resp.StatusCode == 200 {
			var clusterInfo ElasticsearchClusterInfo
			err = json.NewDecoder(resp.Body).Decode(&clusterInfo)
			if err != nil {
				return resource.NonRetryableError(err)
			}

			if clusterInfo.Status == status {
				log.Printf("[DEBUG] WaitForElasticsearchClusterStatus desired cluster status reached: %s\n", clusterInfo.Status)
				return nil
			}

			log.Printf("[DEBUG] WaitForElasticsearchClusterStatus current cluster status: %s. Desired status: %s\n", clusterInfo.Status, status)
		}

		return resource.RetryableError(
			fmt.Errorf("%q: timeout while waiting for the elasticsearch cluster to reach %s status", id, status))
	})
}

// WaitForKibanaClusterStatus waits for a Kibana cluster to enter the specified status.
func (c *ECEClient) WaitForKibanaClusterStatus(id string, status string, allowMissing bool) error {
	timeoutSeconds := time.Second * time.Duration(c.Timeout)
	log.Printf("[DEBUG] WaitForKibanaClusterStatus will wait for %v seconds for '%s' status for Kibana cluster ID: %s\n", timeoutSeconds, status, id)

	return resource.Retry(timeoutSeconds, func() *resource.RetryError {
		resp, err := c.GetKibanaCluster(id)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		if resp.StatusCode == 404 && allowMissing {
			return nil
		} else if resp.StatusCode == 200 {
			var clusterInfo KibanaClusterInfo
			err = json.NewDecoder(resp.Body).Decode(&clusterInfo)
			if err != nil {
				return resource.NonRetryableError(err)
			}

			if clusterInfo.Status == status {
				log.Printf("[DEBUG] WaitForKibanaClusterStatus desired Kibana cluster status reached: %s\n", clusterInfo.Status)
				return nil
			}

			log.Printf("[DEBUG] WaitForKibanaClusterStatus current Kibana cluster status: %s. Desired status: %s\n", clusterInfo.Status, status)
		}

		return resource.RetryableError(
			fmt.Errorf("%q: timeout while waiting for the Kibana cluster to reach %s status", id, status))
	})
}
