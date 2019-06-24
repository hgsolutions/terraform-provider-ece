package main

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceServerCreate,
		Read:   resourceServerRead,
		Update: resourceServerUpdate,
		Delete: resourceServerDelete,

		Schema: map[string]*schema.Schema{
			"ece_endpoint": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The fully-qualified endpoint for Elastic ECE, including port.",
			},
			"ece_user": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The username to connect using basic auth.",
				DefaultFunc: schema.EnvDefaultFunc("ECE_USER", nil),
			},
			"ece_pass": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The password to connect using basic auth.",
				DefaultFunc: schema.EnvDefaultFunc("ECE_PASS", nil),
				Sensitive:   true,
			},
		},
	}
}

func resourceServerCreate(d *schema.ResourceData, m interface{}) error {
	return resourceServerRead(d, m)
}

func resourceServerRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceServerUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceServerRead(d, m)
}

func resourceServerDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
