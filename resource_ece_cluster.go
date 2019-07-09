package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
				Type:        schema.TypeString,
				ForceNew:    false,
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

	clusterPlan, err := expandClusterPlan(d, meta)
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

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Cluster response body: %v\n", string(respBytes))

	var clusterInfo ElasticsearchClusterInfo
	err = json.Unmarshal(respBytes, &clusterInfo)
	if err != nil {
		return err
	}

	d.Set("name", clusterInfo.ClusterName)

	currentPlan := clusterInfo.PlanInfo.Current.Plan

	err = d.Set("elasticsearch_version", currentPlan.Elasticsearch.Version)
	if err != nil {
		return err
	}

	// NOTE: This property appears as deprecated in the ECE API documentation, recommending use of the zone count from the
	// ElasticsearchClusterTopologyElement instead. However, zone count is not returned for ElasticsearchClusterTopologyElement
	// in the current version of ECE (2.2.3). To support either location, the zone count is used from cluster plan unless the
	// cluster topology element has a non-zero value.
	// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ElasticsearchClusterPlan
	zoneCount := currentPlan.ZoneCount

	if len(currentPlan.ClusterTopology) > 0 {
		clusterTopology := currentPlan.ClusterTopology[0]

		err = d.Set("memory_per_node", clusterTopology.MemoryPerNode)
		if err != nil {
			return err
		}

		err = d.Set("node_count_per_zone", clusterTopology.NodeCountPerZone)
		if err != nil {
			return err
		}

		// See note above about clusterPlan.ZoneCount.
		if clusterTopology.ZoneCount > 0 {
			zoneCount = clusterTopology.ZoneCount
		}
	}

	err = d.Set("zone_count", zoneCount)
	if err != nil {
		return err
	}

	if len(clusterInfo.Topology.Instances) > 0 {
		instance := clusterInfo.Topology.Instances[0]

		nodeType := &ElasticsearchNodeType{}

		if instance.ServiceRoles != nil {
			nodeTypeMap := make(map[string]interface{}, len(instance.ServiceRoles))
			for _, s := range instance.ServiceRoles {
				nodeTypeMap[s] = true
			}

			expandNodeTypeFromMap(nodeType, nodeTypeMap)
		}

		flatNodeType := flattenNodeType(nodeType)

		err = d.Set("node_type", flatNodeType)
		if err != nil {
			return err
		}
	}

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
		return fmt.Errorf("%q: cluster ID was not found for update", clusterID)
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

	clusterPlan, err := expandClusterPlan(d, meta)
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

	// Confirm that the update plan was successfully applied.
	resp, err = client.GetClusterPlanActivity(clusterID)
	if err != nil {
		return err
	}

	if resp.StatusCode == 404 {
		return fmt.Errorf("%q: cluster ID was not found after update", clusterID)
	}

	var clusterPlansInfo ElasticsearchClusterPlansInfo
	err = json.NewDecoder(resp.Body).Decode(&clusterPlansInfo)
	if err != nil {
		return err
	}

	if !clusterPlansInfo.Current.Healthy {
		var logMessages interface{}
		failedLogMessages := make([]ClusterPlanStepLogMessageInfo, 0)
		// Attempt to find the failed step in the plan.
		if clusterPlansInfo.Current.PlanAttemptLog != nil {
			for _, stepInfo := range clusterPlansInfo.Current.PlanAttemptLog {
				if stepInfo.Status != "success" {
					for _, logMessageInfo := range stepInfo.InfoLog {
						failedLogMessages = append(failedLogMessages, logMessageInfo)
					}
				}
			}
		}

		logMessages, err := json.MarshalIndent(failedLogMessages, "", " ")
		if err != nil {
			log.Printf("[DEBUG] Error marshalling log messages to JSON: %v\n", err)

			logMessages = failedLogMessages
		} else {
			logMessages = string(logMessages.([]byte))
		}

		return fmt.Errorf("%q: cluster update failed: %v", clusterID, logMessages)
	}

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

func expandClusterPlan(d *schema.ResourceData, meta interface{}) (clusterPlan *ElasticsearchClusterPlan, err error) {
	//client := meta.(*ECEClient)

	nodeType, err := expandNodeType(d, meta)
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
		// Commenting because the default is calculated based on cluster size and is
		// typically higher than the configured provider timeout.
		// Transient: TransientElasticsearchPlanConfiguration{
		// 	PlanConfiguration: ElasticsearchPlanControlConfiguration{
		// 		Timeout: int64(client.timeout),
		// 	},
		// },
	}

	return clusterPlan, nil
}

func expandNodeType(d *schema.ResourceData, meta interface{}) (nodeType *ElasticsearchNodeType, err error) {
	nodeType = DefaultElasticsearchNodeType()

	if v, ok := d.GetOk("node_type"); ok {
		nodeTypeList := v.(*schema.Set).List()
		for _, vv := range nodeTypeList {
			nodeTypeMap := vv.(map[string]interface{})
			expandNodeTypeFromMap(nodeType, nodeTypeMap)
		}
	}

	return nodeType, nil
}

func expandNodeTypeFromMap(nodeType *ElasticsearchNodeType, nodeTypeMap map[string]interface{}) {
	if v, ok := nodeTypeMap["data"]; ok {
		nodeType.Data = v.(bool)
		log.Printf("[DEBUG] Expanded node_type.data as: %t\n", nodeType.Data)
	}

	if v, ok := nodeTypeMap["ingest"]; ok {
		nodeType.Ingest = v.(bool)
		log.Printf("[DEBUG] Expanded node_type.ingest as: %t\n", nodeType.Ingest)
	}

	if v, ok := nodeTypeMap["master"]; ok {
		nodeType.Master = v.(bool)
		log.Printf("[DEBUG] Expanded node_type.master as: %t\n", nodeType.Master)
	}

	if v, ok := nodeTypeMap["ml"]; ok {
		nodeType.ML = v.(bool)
		log.Printf("[DEBUG] Expanded node_type.ml as: %t\n", nodeType.ML)
	}
}

func flattenNodeType(nodeType *ElasticsearchNodeType) []map[string]interface{} {
	m := make([]map[string]interface{}, 0)

	mm := map[string]interface{}{}

	mm["data"] = nodeType.Data
	log.Printf("[DEBUG] Flattened node_type.data as: %t\n", mm["data"])

	mm["ingest"] = nodeType.Ingest
	log.Printf("[DEBUG] Flattened node_type.ingest as: %t\n", mm["ingest"])

	mm["master"] = nodeType.Master
	log.Printf("[DEBUG] Flattened node_type.master as: %t\n", mm["master"])

	mm["ml"] = nodeType.ML
	log.Printf("[DEBUG] Flattened node_type.ml as: %t\n", mm["ml"])

	m = append(m, mm)

	return m
}
