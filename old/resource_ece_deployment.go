package main

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceEceDeployment() *schema.Resource {
	return &schema.Resource{
		Create: resourceEceDeploymentCreate,
		Read:   resourceEceDeploymentRead,
		Update: resourceEceDeploymentUpdate,
		Delete: resourceEceDeploymentDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"settings": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
			},
		},
	}
}

func resourceEceDeploymentCreate(d *schema.ResourceData, m interface{}) error {

	return resourceEceDeploymentRead(d, m)
}

func resourceEceDeploymentRead(d *schema.ResourceData, m interface{}) error {

	return nil
}

func resourceEceDeploymentUpdate(d *schema.ResourceData, m interface{}) error {

	return resourceEceDeploymentRead(d, m)
}

func resourceEceDeploymentDelete(d *schema.ResourceData, m interface{}) error {
	var err error

	return err
}
