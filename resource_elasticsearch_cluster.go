package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceECECluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceECEClusterCreate,
		Read:   resourceECEClusterRead,
		Update: resourceECEClusterUpdate,
		Delete: resourceECEClusterDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type: schema.TypeString,
				// Unsure if we need this set to true here. If a change occurs in the body/json, can an update happen or does
				// the cluster need to be deleted and recreated?
				ForceNew:    false, // https://github.com/hashicorp/terraform/blob/master/helper/schema/schema.go#L184
				Required:    true,
				Description: "The name of the cluster",
			},
			"elasticsearch_version": &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    false,
				Required:    true,
				Description: "The version of the Elasticsearch cluster (must be one of the ECE supported versions).",
			},
			"memory_per_node": &schema.Schema{
				Type:        schema.TypeInt,
				ForceNew:    false,
				Optional:    true,
				Default:     2048,
				Description: "The memory capacity in MB for each node of this type built in each zone. The default is 2048.",
			},
			"node_count_per_zone": &schema.Schema{
				Type:        schema.TypeInt,
				ForceNew:    false,
				Optional:    true,
				Default:     1,
				Description: "The number of nodes of this type that are allocated within each zone. The default is 1.",
			},
			"node_type": {
				Type:        schema.TypeSet,
				Description: "Controls the combinations of Elasticsearch node types. By default, the Elasticsearch node is master eligible, can hold data, and run ingest pipelines.",
				ForceNew:    false,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Defines whether this node can hold data. The default is true.",
						},
						"ingest": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Defines whether this node can run an ingest pipeline. The default is true.",
						},
						"master": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Defines whether this node can be elected master. The default is true.",
						},
						"ml": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Defines whether this node can run ml jobs, valid only for versions 5.4.0 or greater. The default is false.",
						},
					},
				},
			},
			"zone_count": &schema.Schema{
				Type:        schema.TypeInt,
				ForceNew:    false,
				Optional:    true,
				Default:     1,
				Description: "The default number of zones in which data nodes will be placed. The default is 1.",
			},
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceECEClusterCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ECEClient)

	// TODO: Determine whether the named cluster already exists...

	data := true
	ingest := true
	master := true
	ml := false

	// nodeTypes := d.Get("node_type").(*schema.Set)
	// for _, raw := range nodeTypes {
	// 	t := raw.(map[string]interface{})
	// 	if val, ok := t["data"]; ok {
	// 		data := t["data"].(bool)
	// 	}

	// }

	// Example cluster creation request body.
	// {
	// 	"cluster_name" : "My First Deployment",
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
	// 	}
	//   }

	createClusterRequest := CreateElasticsearchClusterRequest{
		ClusterName: d.Get("name").(string),
		Plan: ElasticsearchClusterPlan{
			Elasticsearch: ElasticsearchConfiguration{
				Version: d.Get("elasticsearch_version").(string),
			},
			ClusterTopology: []ElasticsearchClusterTopologyElement{
				ElasticsearchClusterTopologyElement{
					MemoryPerNode:    d.Get("memory_per_node").(int),
					NodeCountPerZone: d.Get("node_count_per_zone").(int),
					NodeType: ElasticsearchNodeType{
						Data:   data,
						Ingest: ingest,
						Master: master,
						ML:     ml,
					},
					ZoneCount: d.Get("zone_count").(int),
				},
			},
		},
	}

	jsonData, err := json.Marshal(createClusterRequest)
	if err != nil {
		return err
	}

	jsonString := string(jsonData)
	log.Printf("[DEBUG] JSON Request: %v\n", jsonString)

	resp, err := client.CreateCluster(jsonString)
	if err != nil {
		return err
	}

	// Example response:
	// {
	// 	"elasticsearch_cluster_id": "5de00f3876e3442f8e4f83110af0e251",
	// 	"credentials": {
	// 		"username": "elastic",
	// 		"password": "Ov8cmAVCqTr8biFfND2wtIuY"
	// 	}
	// }

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] CreateCluster response body: %v\n", string(respBytes))

	var crudResponse ClusterCrudResponse
	err = json.Unmarshal(respBytes, &crudResponse)
	if err != nil {
		return err
	}

	clusterID := crudResponse.ElasticsearchClusterID
	log.Printf("[DEBUG] Created cluster ID: %s\n", clusterID)

	err = client.WaitForCreate(clusterID)
	if err != nil {
		return err
	}

	d.SetId(clusterID)

	return resourceECEClusterRead(d, meta)
}

func resourceECEClusterRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ECEClient)

	clusterID := d.Id()
	log.Printf("[DEBUG] Reading cluster plan for cluster ID: %s\n", clusterID)

	resp, err := client.GetCluster(clusterID)
	if err != nil {
		return err
	}

	if resp.StatusCode == 404 {
		return fmt.Errorf("[WARN] Cluster ID was not found: %+v", resp)
	}

	jsonBody, err := client.GetResponseAsJSON(resp)
	if err != nil {
		return err
	}

	d.Set("cluster", jsonBody)

	return nil
}

func resourceECEClusterUpdate(d *schema.ResourceData, meta interface{}) error {

	return resourceECEClusterRead(d, meta)
}

func resourceECEClusterDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ECEClient)
	clusterID := d.Id()

	// NOTE: A cluster must be successfully _shutdown first before it can be deleted.
	log.Printf("[DEBUG] Shutting down cluster ID: %s\n", clusterID)
	_, err := client.ShutdownCluster(clusterID)
	if err != nil {
		return err
	}

	// Wait for cluster shutdown.
	log.Printf("[DEBUG] Waiting for shutdown of cluster ID: %s\n", clusterID)
	client.WaitForShutdown(clusterID)

	log.Printf("[DEBUG] Deleting cluster ID: %s\n", clusterID)
	_, err = client.DeleteCluster(clusterID)
	if err != nil {
		return err
	}

	return nil
}
