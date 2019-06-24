package main

/*
Modeled after structure & functionality as found here: https://github.com/phillbaker/terraform-provider-elasticsearch
*/

import (
	"encoding/base64"
	"net/http"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

/*
Example

provider "ece" {
	ece_endpoint	= "http://ec2-3-86-31-57.compute-1.amazonaws.com:12400"
	ece_user		= "my_user"
	ece_pass		= "my_pass"
}

*/

// Basic ECE needs.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"ece_endpoint": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ECE_ENDPOINT", nil),
				Description: "The fully-qualified endpoint for ECE communications, including port.",
			},
			"ece_user": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ECE_USER", nil),
				Description: "The username to connect using basic authentication.",
			},
			"ece_pass": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ECE_PASS", nil),
				Description: "The password to connect using basic authentication.",
				Sensitive:   true,
			}, /*
				"cacert_file": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "A Custom CA certificate",
				},
				"insecure": &schema.Schema{
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "Disable SSL verification of API calls",
				},*/
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	var relevantClient interface{}
	relevantClient = getHttpClient(d)
	return relevantClient, nil
}

// HTTP client.
func getHttpClient(d *schema.ResourceData) *http.Client {
	// *** These need to be available provider-wide vs. just in this .go.
	//endpoint := d.Get("ece_endpoint").(string)
	//user := d.Get("ece_user").(string)
	//pass := d.Get("ece_pass").(string)

	client := &http.Client{}
	//CheckRedirect: redirectPolicyFunc,
	return client
}

// This might need to be relocated to a util.go or something (used in other places also).
func getBasicAuthHeader(user, pass string) string {
	auth := user + ":" + pass
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// Might need to support redirects?
//func redirectPolicyFunc(req *http.Request, via []*http.Request) error {
//	req.Header.Add("Authorization", "Basic "+getBasicAuth())
//}

/*
// HTTPS client.
func tlsHttpClient(d *schema.ResourceData) *http.Client {
	insecure := d.Get("insecure").(bool)
	cacertFile := d.Get("cacert_file").(string)

	// Configure TLS/SSL
	tlsConfig := &tls.Config{}

	// If a cacertFile has been specified, use that for cert validation
	if cacertFile != "" {
		caCert, _, _ := pathorcontents.Read(cacertFile)

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(caCert))
		tlsConfig.RootCAs = caCertPool
	}

	// If configured as insecure, turn off SSL verification
	if insecure {
		tlsConfig.InsecureSkipVerify = true
	}

	transport := &http.Transport{TLSClientConfig: tlsConfig}

	client := &http.Client{Transport: transport}

	return client
}
*/
