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
	// NOTE: Several of the aggregate schema resources below would better be mapped as TypeMap,
	// but currently TypeMap cannot be used for non-string values due to this bug:
	// https://github.com/hashicorp/terraform/issues/15327
	// As a result, I used TypeList with a MaxValue of 1, matching what is done with the AWS
	// provider for Elasticsearch domains. See the following for examples:
	// github.com/terraform-providers/terraform-provider-aws/aws/resource_aws_elasticsearch_domain.go

	return &schema.Resource{
		Create: resourceECEClusterCreate,
		Read:   resourceECEClusterRead,
		Update: resourceECEClusterUpdate,
		Delete: resourceECEClusterDelete,
		Schema: map[string]*schema.Schema{
			"cluster_name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The name of the cluster.",
				ForceNew:    false,
				Required:    true,
			},
			"plan": {
				Type:        schema.TypeList,
				Description: "The plan for the Elasticsearch cluster.",
				ForceNew:    false,
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cluster_topology": {
							Type:        schema.TypeList,
							Description: "The topology of the Elasticsearch nodes, including the number, capacity, and type of nodes, and where they can be allocated.",
							Optional:    true,
							Computed:    false,
							MinItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"memory_per_node": &schema.Schema{
										Type:        schema.TypeInt,
										Description: "The memory capacity in MB for each node of this type built in each zone. The default is 2048.",
										ForceNew:    false,
										Optional:    true,
										Default:     1024,
									},
									"node_count_per_zone": &schema.Schema{
										Type:        schema.TypeInt,
										Description: "The number of nodes of this type that are allocated within each zone. The default is 1.",
										ForceNew:    false,
										Optional:    true,
										Default:     1,
									},
									"node_type": {
										Type:        schema.TypeList,
										Description: "Controls the combinations of Elasticsearch node types. By default, the Elasticsearch node is master eligible, can hold data, and run ingest pipelines.",
										ForceNew:    false,
										Optional:    true,
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"data": {
													Type:        schema.TypeBool,
													Description: "Defines whether this node can hold data. The default is true.",
													Optional:    true,
													Default:     true,
												},
												"ingest": {
													Type:        schema.TypeBool,
													Description: "Defines whether this node can run an ingest pipeline. The default is true.",
													Optional:    true,
													Default:     true,
												},
												"master": {
													Type:        schema.TypeBool,
													Description: "Defines whether this node can be elected master. The default is true.",
													Optional:    true,
													Default:     true,
												},
												"ml": {
													Type:        schema.TypeBool,
													Description: "Defines whether this node can run ml jobs, valid only for versions 5.4.0 or greater. Not supported in OSS ECE. The default is false.",
													Optional:    true,
													Default:     false,
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
							},
						},
						"elasticsearch": {
							Type:        schema.TypeList,
							Description: "The Elasticsearch cluster settings.",
							ForceNew:    false,
							Required:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"version": &schema.Schema{
										Type:        schema.TypeString,
										Description: "The version of the Elasticsearch cluster (must be one of the ECE supported versions).",
										ForceNew:    false,
										Required:    true,
									},
								},
							},
						},
					},
				},
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

	clusterName := d.Get("cluster_name").(string)
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

	log.Printf("[DEBUG] Setting cluster_name: %v\n", clusterInfo.ClusterName)
	d.Set("cluster_name", clusterInfo.ClusterName)

	plan := flattenClusterPlan(clusterInfo)
	log.Printf("[DEBUG] Setting cluster plan: %v\n", plan)
	d.Set("plan", plan)
	if err != nil {
		return err
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

	if d.HasChange("cluster_name") {
		metadata := ClusterMetadataSettings{
			ClusterName: d.Get("cluster_name").(string),
		}

		_, err = client.UpdateClusterMetadata(clusterID, metadata)
		if err != nil {
			return err
		}
	}

	d.SetPartial("cluster_name")

	if d.HasChange("plan") {
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
	clusterPlanList := d.Get("plan").([]interface{})
	clusterPlanMap := clusterPlanList[0].(map[string]interface{})

	clusterTopology := expandClusterTopology(clusterPlanMap)
	elasticsearchConfiguration, err := expandElasticsearchConfiguration(clusterPlanMap)
	if err != nil {
		return nil, err
	}

	clusterPlan = &ElasticsearchClusterPlan{
		Elasticsearch:   *elasticsearchConfiguration,
		ClusterTopology: clusterTopology,
	}

	return clusterPlan, nil
}

func expandClusterTopology(clusterPlanMap map[string]interface{}) []ElasticsearchClusterTopologyElement {
	inputClusterTopologyMap := clusterPlanMap["cluster_topology"].([]interface{})
	clusterTopology := make([]ElasticsearchClusterTopologyElement, 0)

	for _, t := range inputClusterTopologyMap {
		elementMap := t.(map[string]interface{})
		clusterTopologyElement := DefaultElasticsearchClusterTopologyElement()
		if v, ok := elementMap["memory_per_node"]; ok {
			clusterTopologyElement.MemoryPerNode = v.(int)
		}

		if v, ok := elementMap["node_count_per_zone"]; ok {
			clusterTopologyElement.NodeCountPerZone = v.(int)
		}

		if v, ok := elementMap["node_type"]; ok {
			nodeType := DefaultElasticsearchNodeType()
			nodeTypeMaps := v.([]interface{})
			if len(nodeTypeMaps) > 0 {
				expandNodeTypeFromMap(nodeType, nodeTypeMaps[0].(map[string]interface{}))
			}
			clusterTopologyElement.NodeType = *nodeType
		}

		if v, ok := elementMap["zone_count"]; ok {
			clusterTopologyElement.ZoneCount = v.(int)
		}

		clusterTopology = append(clusterTopology, *clusterTopologyElement)
	}

	// Create a default cluster topology element if none is provided in the input map.
	if len(clusterTopology) == 0 {
		clusterTopology = append(clusterTopology, *DefaultElasticsearchClusterTopologyElement())
	}

	return clusterTopology
}

func expandElasticsearchConfiguration(clusterPlanMap map[string]interface{}) (elasticsearchConfiguration *ElasticsearchConfiguration, err error) {
	// Get the single elasticsearch element from the plan element.
	elasticsearchList := clusterPlanMap["elasticsearch"].([]interface{})

	if len(elasticsearchList) < 1 {
		return nil, fmt.Errorf("cluster version is required")
	}

	elasticsearchMap, ok := elasticsearchList[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("cluster version is required")
	}

	elasticsearchConfiguration = &ElasticsearchConfiguration{
		Version: elasticsearchMap["version"].(string),
	}

	return elasticsearchConfiguration, nil
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

func flattenClusterPlan(clusterInfo ElasticsearchClusterInfo) []map[string]interface{} {
	clusterPlanMaps := make([]map[string]interface{}, 1)

	clusterPlan := clusterInfo.PlanInfo.Current.Plan

	clusterPlanMap := make(map[string]interface{})
	clusterPlanMap["cluster_topology"] = flattenClusterTopology(clusterInfo, clusterPlan)
	clusterPlanMap["elasticsearch"] = flattenElasticsearchConfiguration(clusterPlan.Elasticsearch)

	clusterPlanMaps[0] = clusterPlanMap

	return clusterPlanMaps
}

func flattenClusterTopology(clusterInfo ElasticsearchClusterInfo, clusterPlan ElasticsearchClusterPlan) []map[string]interface{} {
	topologyMap := make([]map[string]interface{}, 0)

	// NOTE: This property appears as deprecated in the ECE API documentation, recommending use of the zone count from the
	// ElasticsearchClusterTopologyElement instead. However, zone count is not returned for ElasticsearchClusterTopologyElement
	// in the current version of ECE (2.2.3). To support either location, the zone count is used from cluster plan unless the
	// cluster topology element has a non-zero value.
	// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ElasticsearchClusterPlan
	defaultZoneCount := clusterPlan.ZoneCount

	for i, t := range clusterPlan.ClusterTopology {
		elementMap := make(map[string]interface{})

		elementMap["memory_per_node"] = t.MemoryPerNode
		elementMap["node_count_per_zone"] = t.NodeCountPerZone

		elementMap["node_type"] = flattenNodeType(clusterInfo, i)

		// See note above about clusterPlan.ZoneCount.
		if t.ZoneCount > 0 {
			elementMap["zone_count"] = t.ZoneCount
		} else {
			elementMap["zone_count"] = defaultZoneCount
		}

		topologyMap = append(topologyMap, elementMap)
	}

	logJSON("Flattened cluster topology", topologyMap)

	return topologyMap
}

func flattenElasticsearchConfiguration(configuration ElasticsearchConfiguration) []map[string]interface{} {
	elasticsearchMaps := make([]map[string]interface{}, 1)

	elasticsearchMap := make(map[string]interface{})
	elasticsearchMap["version"] = configuration.Version

	elasticsearchMaps[0] = elasticsearchMap

	logJSON("Flattened elasticsearch configuration", elasticsearchMaps)

	return elasticsearchMaps
}

func flattenNodeType(clusterInfo ElasticsearchClusterInfo, instanceIndex int) map[string]interface{} {
	nodeTypeMap := make(map[string]interface{})

	if len(clusterInfo.Topology.Instances) > 0 {
		instance := clusterInfo.Topology.Instances[instanceIndex]

		nodeType := &ElasticsearchNodeType{}

		if instance.ServiceRoles != nil {
			nodeTypeValues := make(map[string]interface{})
			for _, s := range instance.ServiceRoles {
				nodeTypeValues[s] = true
			}

			expandNodeTypeFromMap(nodeType, nodeTypeValues)
		}

		nodeTypeMap["data"] = nodeType.Data
		log.Printf("[DEBUG] Flattened node_type.data as: %t\n", nodeTypeMap["data"])

		nodeTypeMap["ingest"] = nodeType.Ingest
		log.Printf("[DEBUG] Flattened node_type.ingest as: %t\n", nodeTypeMap["ingest"])

		nodeTypeMap["master"] = nodeType.Master
		log.Printf("[DEBUG] Flattened node_type.master as: %t\n", nodeTypeMap["master"])

		nodeTypeMap["ml"] = nodeType.ML
		log.Printf("[DEBUG] Flattened node_type.ml as: %t\n", nodeTypeMap["ml"])
	}

	return nodeTypeMap
}

func logJSON(context string, m interface{}) {
	jsonBytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		log.Printf("[DEBUG] %s: error marshalling value as JSON: %s. %v", context, err, m)
	}

	log.Printf("[DEBUG] %s: %s", context, string(jsonBytes))
}
