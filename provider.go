package main

/*
Modeled after structure & functionality as found here: https://github.com/phillbaker/terraform-provider-elasticsearch
*/

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

/*
Example

provider "ece" {
    url      = "http://ece-api-url:12400"
    username = ""
    password = ""
    insecure = true # to bypass certificate check
}

*/

// Provider for ECE cluster management using Terraform.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ECE_URL", nil),
				Description: "The fully-qualified URL for the ECE API, including port.",
			},
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ECE_USERNAME", nil),
				Description: "The ECE username to use for basic authentication.",
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ECE_PASSWORD", nil),
				Description: "The ECE password to use for basic authentication.",
				Sensitive:   true,
			},
			"timeout": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     3600,
				Description: "The timeout in seconds for resource operations. The default is 1 hour (3600 seconds).",
			},
			"insecure": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Disable SSL verification of API calls.",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"ece_cluster": resourceECECluster(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	rawURL := d.Get("url").(string)
	log.Printf("[DEBUG] Connecting to ECE: %s\n", rawURL)

	_, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	username := d.Get("username").(string)
	password := d.Get("password").(string)
	timeout := d.Get("timeout").(int)

	log.Printf("[DEBUG] ECE username: %s\n", username)
	//log.Printf("[DEBUG] ECE password: %s\n", password)
	log.Printf("[DEBUG] ECE timeout: %v\n", timeout)

	httpClient := getHTTPClient(d)

	eceClient := &ECEClient{
		httpClient: httpClient,
		url:        rawURL,
		username:   username,
		password:   password,
		timeout:    timeout,
	}

	return eceClient, nil
}

func getHTTPClient(d *schema.ResourceData) *http.Client {
	insecure := d.Get("insecure").(bool)
	timeout := d.Get("timeout").(int)

	// Configure TLS/SSL
	tlsConfig := &tls.Config{}

	// If configured as insecure, turn off SSL verification
	if insecure {
		tlsConfig.InsecureSkipVerify = true
	}

	transport := &http.Transport{TLSClientConfig: tlsConfig}

	client := &http.Client{Transport: transport}

	client.Timeout = time.Second * time.Duration(timeout)

	log.Printf("[DEBUG] HTTP client timeout: %v\n", client.Timeout)

	return client
}
