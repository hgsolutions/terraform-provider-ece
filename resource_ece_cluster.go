package main

import (
	"fmt"
	"log"
	"time"

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
				MaxItems:    1,
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
							Default:     false,
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
			"elasticsearch_username": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The username for the created cluster.",
			},
			"elasticsearch_password": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The password for the created cluster.",
			},
		},
		// TODO: Test import of existing clusters
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceECEClusterCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ECEClient)

	// TODO: Consider whether any other settings are required for v1 of the provider. Kibana cluster?

	clusterName := d.Get("name").(string)
	log.Printf("[DEBUG] Creating cluster with name: %s\n", clusterName)

	clusterPlan, err := buildClusterPlan(d, meta)
	if err != nil {
		return err
	}

	createClusterRequest := CreateElasticsearchClusterRequest{
		ClusterName: clusterName,
		Plan:        *clusterPlan,
	}

	crudResponse, err := client.CreateCluster(createClusterRequest)
	if err != nil {
		return err
	}

	clusterID := crudResponse.ElasticsearchClusterID
	log.Printf("[DEBUG] Created cluster ID: %s\n", clusterID)

	err = client.WaitForStatus(clusterID, "started")
	if err != nil {
		return err
	}

	d.SetId(clusterID)
	d.Set("elasticsearch_username", crudResponse.Credentials.Username)
	d.Set("elasticsearch_password", crudResponse.Credentials.Password)

	return resourceECEClusterRead(d, meta)
}

func resourceECEClusterRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ECEClient)

	clusterID := d.Id()
	log.Printf("[DEBUG] Reading cluster information for cluster ID: %s\n", clusterID)

	resp, err := client.GetCluster(clusterID)
	if err != nil {
		return err
	}

	// If the resource does not exist, inform Terraform. We want to immediately
	// return here to prevent further processing.
	if resp.StatusCode == 404 {
		log.Printf("[DEBUG] cluster ID not found: %s\n", clusterID)
		d.SetId("")
		return nil
	}

	jsonBody, err := client.GetResponseBodyAsJSON(resp)
	if err != nil {
		return err
	}

	d.Set("cluster", jsonBody)

	return nil
}

func resourceECEClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ECEClient)

	d.Partial(true)

	clusterID := d.Id()
	log.Printf("[DEBUG] Updating cluster ID: %s\n", clusterID)

	resp, err := client.GetCluster(clusterID)
	if err != nil {
		return err
	}

	if resp.StatusCode == 404 {
		return fmt.Errorf("%q: Cluster ID was not found: ", clusterID)
	}

	if d.HasChange("name") {
		metadata := ClusterMetadataSettings{
			ClusterName: d.Get("name").(string),
		}

		_, err = client.UpdateClusterMetadata(clusterID, metadata)
		if err != nil {
			return err
		}
	}

	d.SetPartial("name")

	clusterPlan, err := buildClusterPlan(d, meta)
	if err != nil {
		return err
	}

	_, err = client.UpdateCluster(clusterID, *clusterPlan)
	if err != nil {
		return err
	}

	// Wait for the cluster plan to be initiated.
	duration := time.Duration(5) * time.Second // 5 seconds
	time.Sleep(duration)

	err = client.WaitForStatus(clusterID, "started")
	if err != nil {
		return err
	}

	// TODO: A plan may fail to update the cluster even if the update is accepted. Get the latest cluster
	// plan and ensure it matches the desired plan before indicating success of the update.

	d.Partial(false)

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

func buildClusterPlan(d *schema.ResourceData, meta interface{}) (clusterPlan *ElasticsearchClusterPlan, err error) {
	nodeType, err := buildNodeType(d, meta)
	if err != nil {
		return nil, err
	}

	clusterPlan = &ElasticsearchClusterPlan{
		Elasticsearch: ElasticsearchConfiguration{
			Version: d.Get("elasticsearch_version").(string),
		},
		ClusterTopology: []ElasticsearchClusterTopologyElement{
			ElasticsearchClusterTopologyElement{
				MemoryPerNode:    d.Get("memory_per_node").(int),
				NodeCountPerZone: d.Get("node_count_per_zone").(int),
				NodeType:         *nodeType,
				ZoneCount:        d.Get("zone_count").(int),
			},
		},
	}

	return clusterPlan, nil
}

func buildNodeType(d *schema.ResourceData, meta interface{}) (nodeType *ElasticsearchNodeType, err error) {
	nodeType = &ElasticsearchNodeType{
		Data:   true,
		Ingest: true,
		Master: true,
		ML:     false,
	}

	if v, ok := d.GetOk("node_type"); ok {
		nodeTypeList := v.(*schema.Set).List()
		for _, vv := range nodeTypeList {
			nt := vv.(map[string]interface{})

			if v, ok := nt["data"]; ok {
				nodeType.Data = v.(bool)
				log.Printf("[DEBUG] Setting node_type.data: %t\n", nodeType.Data)
			}
			if v, ok := nt["ingest"]; ok {
				nodeType.Ingest = v.(bool)
				log.Printf("[DEBUG] Setting node_type.ingest: %t\n", nodeType.Ingest)
			}
			if v, ok := nt["master"]; ok {
				nodeType.Master = v.(bool)
				log.Printf("[DEBUG] Setting node_type.master: %t\n", nodeType.Master)
			}
			if v, ok := nt["ml"]; ok {
				nodeType.ML = v.(bool)
				log.Printf("[DEBUG] Setting node_type.ml: %t\n", nodeType.ML)
			}
		}
	}

	return nodeType, nil
}
