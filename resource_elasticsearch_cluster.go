package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
)

const eceResource = "/api/v1/clusters/elasticsearch"

func resourceElasticsearchCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchClusterCreate,
		Read:   resourceElasticsearchClusterRead,
		Update: resourceElasticsearchClusterUpdate,
		Delete: resourceElasticsearchClusterDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type: schema.TypeString,
				// Unsure if we need this set to true here. If a change occurs in the body/json, can an update happen or does
				// the cluster need to be deleted and recreated?
				//ForceNew: true, // https://github.com/hashicorp/terraform/blob/master/helper/schema/schema.go#L184
				Required: true,
			},
			"body": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				//DiffSuppressFunc: diffSuppressIndexTemplate, // https://github.com/hashicorp/terraform/blob/master/helper/schema/schema.go#L142
			},
		},
	}
}

func resourceElasticsearchClusterCreate(d *schema.ResourceData, m interface{}) error {

	//return resourceElasticsearchClusterRead(d, m)

	err := resourceElasticsearchClusterPost(d, m)
	if err != nil {
		return err
	}

	d.SetId(d.Get("name").(string)) // ToDo: Check on whether the cluster's ID or name are more applicable here?
	return nil
}

// https://stackoverflow.com/questions/16673766/basic-http-auth-in-go
// https://stackoverflow.com/questions/24455147/how-do-i-send-a-json-string-in-a-post-request-in-go

func resourceElasticsearchClusterPost(d *schema.ResourceData, meta interface{}) error {
	//name := d.Get("name").(string)
	body := d.Get("body").(string)
	client := meta.(*http.Client)

	var url = d.Get("ece_endpoint").(string) + eceResource
	var json = []byte(body)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json))
	req.SetBasicAuth(d.Get("ece_user").(string), d.Get("ece_pass").(string))

	client.Timeout = time.Second * 900 // 15 minute timeout, WARNING: golang http clients never time out!

	resp, err := client.Do(req)

	fmt.Println(resp) // ToDo: Remove

	return err
}

func resourceElasticsearchClusterRead(d *schema.ResourceData, m interface{}) error {

	return nil
}

func resourceElasticsearchClusterUpdate(d *schema.ResourceData, m interface{}) error {

	return resourceElasticsearchClusterRead(d, m)
}

func resourceElasticsearchClusterDelete(d *schema.ResourceData, m interface{}) error {
	var err error

	// NOTE: A cluster must be successfully _shutdown first before it can be deleted.

	return err
}
