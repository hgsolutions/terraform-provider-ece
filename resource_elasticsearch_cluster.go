package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceElasticsearchCluster() *schema.Resource {
	// NOTE: Several of the aggregate schema resources below would better be mapped as TypeMap,
	// but currently TypeMap cannot be used for non-string values due to this bug:
	// https://github.com/hashicorp/terraform/issues/15327
	// As a result, I used TypeList with a MaxValue of 1, matching what is done with the AWS
	// provider for Elasticsearch domains. See the following for examples:
	// github.com/terraform-providers/terraform-provider-aws/aws/resource_aws_elasticsearch_domain.go

	return &schema.Resource{
		Create: resourceDeploymentCreate,
		Read:   resourceDeploymentRead,
		Update: nil,
		Delete: resourceDeploymentDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The name of the cluster.",
				ForceNew:    true,
				Required:    true,
			},
			"region": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The region.",
				ForceNew:    true,
				Optional:    true,
				Default:     "us-east-1",
			},
			"version": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The Elastic-Stack version.",
				ForceNew:    true,
				Optional:    true,
				Default:     "7.6.1",
			},
			"elasticsearch": {
				Type:        schema.TypeList,
				Description: "The plan for a Kibana instance that should be created as part of the Elasticsearch cluster.",
				ForceNew:    true,
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"plan": {
							Type:        schema.TypeList,
							Description: "The plan for the Elasticsearch cluster.",
							ForceNew:    true,
							Required:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cluster_topology": {
										Type:        schema.TypeList,
										Description: "The topology of the Elasticsearch nodes, including the number, capacity, and type of nodes, and where they can be allocated.",
										ForceNew:    true,
										Required:    true,
										Computed:    false,
										MinItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"node_type": {
													Type:        schema.TypeList,
													Description: "Controls the combinations of Elasticsearch node types. By default, the Elasticsearch node is master eligible, can hold data, and run ingest pipelines.",
													ForceNew:    true,
													Optional:    true,
													MaxItems:    1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"data": {
																Type:        schema.TypeBool,
																Description: "Defines whether this node can hold data. The default is true.",
																ForceNew:    true,
																Optional:    true,
																Default:     true,
															},
															"ingest": {
																Type:        schema.TypeBool,
																Description: "Defines whether this node can run an ingest pipeline. The default is true.",
																ForceNew:    true,
																Optional:    true,
																Default:     true,
															},
															"master": {
																Type:        schema.TypeBool,
																Description: "Defines whether this node can be elected master. The default is true.",
																ForceNew:    true,
																Optional:    true,
																Default:     true,
															},
															"ml": {
																Type:        schema.TypeBool,
																Description: "Defines whether this node can run ml jobs, valid only for versions 5.4.0 or greater. Not supported in OSS ECE. The default is false.",
																ForceNew:    true,
																Optional:    true,
																Default:     false,
															},
														},
													},
												},
												"size": {
													Type:        schema.TypeList,
													Description: "Measured by the amount of a resource. The final cluster size is calculated using multipliers from the topology instance configuration.",
													ForceNew:    true,
													Optional:    true,
													MaxItems:    1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"resource": {
																Type:        schema.TypeString,
																Description: "Type of resource.",
																ForceNew:    true,
																Optional:    true,
																Default:     "memory",
															},
															"value": {
																Type:        schema.TypeInt,
																Description: "Amount of resource.",
																ForceNew:    true,
																Optional:    true,
																Default:     1024,
															},
														},
													},
												},
												"zone_count": &schema.Schema{
													Type:        schema.TypeInt,
													ForceNew:    true,
													Optional:    true,
													Default:     1,
													Description: "The default number of zones in which data nodes will be placed. The default is 1.",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"kibana": {
				Type:        schema.TypeList,
				Description: "The plan for a Kibana instance that should be created as part of the Elasticsearch cluster.",
				ForceNew:    true,
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"plan": {
							Type:        schema.TypeList,
							Description: "The plan for the Kibana cluster.",
							ForceNew:    true,
							Required:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cluster_topology": {
										Type:        schema.TypeList,
										Description: "The topology of the Kibana nodes, including the number, capacity, and type of nodes, and where they can be allocated.",
										ForceNew:    true,
										Optional:    true,
										Computed:    false,
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"size": {
													Type:        schema.TypeList,
													Description: "Measured by the amount of a resource. The final cluster size is calculated using multipliers from the topology instance configuration.",
													ForceNew:    true,
													Optional:    true,
													MaxItems:    1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"resource": {
																Type:        schema.TypeString,
																Description: "Type of resource.",
																ForceNew:    true,
																Optional:    true,
																Default:     "memory",
															},
															"value": {
																Type:        schema.TypeInt,
																Description: "Amount of resource.",
																ForceNew:    true,
																Optional:    true,
																Default:     1024,
															},
														},
													},
												},
												"zone_count": &schema.Schema{
													Type:        schema.TypeInt,
													ForceNew:    true,
													Optional:    true,
													Default:     1,
													Description: "The default number of zones in which nodes will be placed. The default is 1.",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"elastic_cloud_id": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Elastic Cloud ID.",
			},
			"deployment_id": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The id of the deployment.",
			},
			"deployment_name": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the deployment.",
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

func resourceDeploymentCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ECEClient)

	deploymentName := d.Get("name").(string)
	log.Printf("[DEBUG] Creating deployment with name: %s\n", deploymentName)
	region := d.Get("region").(string)
	log.Printf("[DEBUG] Creating deployment in region: %s\n", region)
	version := d.Get("version").(string)
	log.Printf("[DEBUG] Creating deployment with Elastic-Stack version: %s\n", version)

	elasticsearchPayload, err := expandElasticsearchPayload(d, meta)
	if err != nil {
		return err
	}

	kibanaPayload, err := expandKibanaPayload(d, meta)
	if err != nil {
		return err
	}

	deploymentCreateResources := DeploymentCreateResources{
		Elasticsearch: []*ElasticsearchPayload{elasticsearchPayload},
		Kibana:        []*KibanaPayload{kibanaPayload},
	}

	deploymentCreateRequest := DeploymentCreateRequest{
		Name:      deploymentName,
		Resources: &deploymentCreateResources,
	}

	createResponse, err := client.CreateDeployment(deploymentCreateRequest)

	d.SetId(createResponse.ID)
	err = d.Set("elastic_cloud_id", createResponse.Resources[0].CloudID)
	if err != nil {
		return err
	}
	err = d.Set("deployment_id", createResponse.ID)
	if err != nil {
		return err
	}
	err = d.Set("deployment_name", createResponse.Name)
	if err != nil {
		return err
	}
	err = d.Set("elasticsearch_username", createResponse.Resources[0].Credentials.Username)
	if err != nil {
		return err
	}
	err = d.Set("elasticsearch_password", createResponse.Resources[0].Credentials.Password)
	if err != nil {
		return err
	}

	return nil
}

func resourceDeploymentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ECEClient)
	deploymentID := d.Id()

	log.Printf("[DEBUG] Deleting deployment ID: %s\n", deploymentID)
	_, err := client.DeleteDeployment(deploymentID)
	if err != nil {
		return err
	}

	return nil
}

func resourceDeploymentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ECEClient)

	deploymentID := d.Id()
	log.Printf("[DEBUG] Reading deployment information for ID: %s\n", deploymentID)

	resp, err := client.GetDeployment(deploymentID)
	if err != nil {
		return err
	}

	// If the resource does not exist, inform Terraform. We want to immediately
	// return here to prevent further processing.
	if resp.StatusCode == 404 {
		log.Printf("[DEBUG] Deployment ID not found: %s\n", deploymentID)
		d.SetId("")
		return nil
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deployment response body: %v\n", string(respBytes))

	var deploymentInfo DeploymentGetResponse
	err = json.Unmarshal(respBytes, &deploymentInfo)
	if err != nil {
		return err
	}

	plan := flattenDeploymentPlan(deploymentInfo)
	log.Printf("[DEBUG] Setting deployment plan: %v\n", plan)
	d.Set("plan", plan)
	if err != nil {
		return err
	}

	return nil
}

func flattenDeploymentPlan(deploymentInfo DeploymentGetResponse) []map[string]interface{} {
	clusterPlanMaps := make([]map[string]interface{}, 1)

	elasticserachClusterInfo := deploymentInfo.Resources.Elasticsearch[0].Info
	deploymentPlan := elasticserachClusterInfo.PlanInfo.Current.Plan

	clusterPlanMap := make(map[string]interface{})
	clusterPlanMap["cluster_topology"] = flattenElasticsearchClusterTopology(*elasticserachClusterInfo, deploymentPlan)
	clusterPlanMap["elasticsearch"] = flattenElasticsearchConfiguration(deploymentPlan.Elasticsearch)

	clusterPlanMaps[0] = clusterPlanMap

	return clusterPlanMaps
}

func expandElasticsearchPayload(d *schema.ResourceData, meta interface{}) (elasticsearchPayload *ElasticsearchPayload, err error) {
	log.Printf("[DEBUG] Expanding ElasticsearchPayload.\n")

	// Get sections from ResourceData.
	region := d.Get("region").(string)
	version := d.Get("version").(string)
	elasticsearch := d.Get("elasticsearch").([]interface{})
	log.Printf("[DEBUG] elasticsearch: %v\n", elasticsearch)
	elasticsearchMap := elasticsearch[0].(map[string]interface{})
	log.Printf("[DEBUG] elasticsearchMap: %v\n", elasticsearchMap)

	elasticsearchPayload = DefaultElasticsearchPayload()
	elasticsearchPayload.Region = region
	elasticsearchPayload.Plan.Elasticsearch.Version = version

	if v, ok := elasticsearchMap["plan"]; ok {
		err := expandElasticsearchClusterPlan(&elasticsearchPayload.Plan, v.(interface{}))
		if err != nil {
			return nil, err
		}
	}

	log.Printf("[DEBUG] ElasticsearchPayload: %v\n", elasticsearchPayload)

	return elasticsearchPayload, nil
}

func expandKibanaPayload(d *schema.ResourceData, meta interface{}) (kibanaPayload *KibanaPayload, err error) {
	log.Printf("[DEBUG] Expanding KibanaPayload.\n")

	// Get sections from ResourceData.
	region := d.Get("region").(string)
	version := d.Get("version").(string)
	kibana := d.Get("kibana").([]interface{})
	kibanaMap := kibana[0].(map[string]interface{})

	kibanaPayload = DefaultKibanaPayload()
	kibanaPayload.Region = region
	kibanaPayload.Plan.Kibana.Version = version

	if v, ok := kibanaMap["plan"]; ok {
		err := expandKibanaClusterPlan(&kibanaPayload.Plan, v.(interface{}))
		if err != nil {
			return nil, err
		}
	}

	log.Printf("[DEBUG] KibanaPayload: %v\n", kibanaPayload)

	return kibanaPayload, nil
}

func expandElasticsearchClusterPlan(elasticsearchPlan *ElasticsearchClusterPlan, inputPlan interface{}) (err error) {
	log.Printf("[DEBUG] Expanding ElasticsearchClusterPlan.\n")

	elasticsearchPlanList := inputPlan.([]interface{})
	elasticsearchPlanMap, ok := elasticsearchPlanList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	expandElasticsearchClusterTopology(elasticsearchPlan, elasticsearchPlanMap)

	return nil
}

func expandElasticsearchClusterTopology(clusterPlan *ElasticsearchClusterPlan, clusterPlanMap map[string]interface{}) {
	log.Printf("[DEBUG] Expanding expandElasticsearchClusterTopology.\n")

	inputClusterTopologyMap := clusterPlanMap["cluster_topology"].([]interface{})
	clusterTopology := make([]ElasticsearchClusterTopologyElement, 0)

	for _, t := range inputClusterTopologyMap {
		elementMap := t.(map[string]interface{})
		clusterTopologyElement := DefaultElasticsearchClusterTopologyElement()

		if v, ok := elementMap["instance_configuration_id"]; ok {
			clusterTopologyElement.InstanceConfigurationID = v.(string)
		}

		log.Printf("[DEBUG] Expanding size.\n")
		if v, ok := elementMap["size"]; ok {
			topologySize := DefaultTopologySize()
			topologySizeMaps := v.([]interface{})
			if len(topologySizeMaps) > 0 {
				expandTopologySizeFromMap(topologySize, topologySizeMaps[0].(map[string]interface{}))
			}
			clusterTopologyElement.Size = *topologySize
		}

		log.Printf("[DEBUG] Expanding node_type.\n")
		if v, ok := elementMap["node_type"]; ok {
			nodeType := DefaultElasticsearchNodeType()
			nodeTypeMaps := v.([]interface{})
			if len(nodeTypeMaps) > 0 {
				expandElasticsearchNodeTypeFromMap(nodeType, nodeTypeMaps[0].(map[string]interface{}))
			}
			clusterTopologyElement.NodeType = *nodeType
		}

		log.Printf("[DEBUG] Expanding zone_count.\n")
		if v, ok := elementMap["zone_count"]; ok {
			clusterTopologyElement.ZoneCount = v.(int)
		}

		clusterTopology = append(clusterTopology, *clusterTopologyElement)
	}

	// Create a default cluster topology element if none is provided in the input map.
	if len(clusterTopology) == 0 {
		clusterTopology = append(clusterTopology, *DefaultElasticsearchClusterTopologyElement())
	}

	clusterPlan.ClusterTopology = clusterTopology

	return
}

func expandKibanaClusterPlan(kibanaPlan *KibanaClusterPlan, inputPlan interface{}) error {
	if inputPlan == nil {
		return nil
	}

	kibanaPlanList := inputPlan.([]interface{})
	if len(kibanaPlanList) == 0 {
		return nil
	}

	kibanaPlanMap, ok := kibanaPlanList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	expandKibanaClusterTopology(kibanaPlan, kibanaPlanMap)

	return nil
}

func expandKibanaClusterTopology(kibanaPlan *KibanaClusterPlan, kibanaPlanMap map[string]interface{}) {
	var inputClusterTopologyMap []interface{}

	if v, ok := kibanaPlanMap["cluster_topology"]; ok {
		inputClusterTopologyMap = v.([]interface{})
	}

	if inputClusterTopologyMap == nil {
		return
	}

	clusterTopology := make([]KibanaClusterTopologyElement, 0)

	for _, t := range inputClusterTopologyMap {
		elementMap := t.(map[string]interface{})
		clusterTopologyElement := DefaultKibanaClusterTopologyElement()

		if v, ok := elementMap["instance_configuration_id"]; ok {
			clusterTopologyElement.InstanceConfigurationID = v.(string)
		}

		if v, ok := elementMap["size"]; ok {
			topologySize := DefaultTopologySize()
			topologySizeMaps := v.([]interface{})
			if len(topologySizeMaps) > 0 {
				expandTopologySizeFromMap(topologySize, topologySizeMaps[0].(map[string]interface{}))
			}
			clusterTopologyElement.Size = *topologySize
		}

		if v, ok := elementMap["zone_count"]; ok {
			clusterTopologyElement.ZoneCount = v.(int)
		}

		clusterTopology = append(clusterTopology, *clusterTopologyElement)
	}

	// Create a default cluster topology element if none is provided in the input map.
	if len(clusterTopology) == 0 {
		clusterTopology = append(clusterTopology, *DefaultKibanaClusterTopologyElement())
	}

	kibanaPlan.ClusterTopology = clusterTopology

	return
}

func expandElasticsearchNodeTypeFromMap(nodeType *ElasticsearchNodeType, nodeTypeMap map[string]interface{}) {
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

func expandTopologySizeFromMap(topologySize *TopologySize, topologySizeMap map[string]interface{}) {
	if v, ok := topologySizeMap["resource"]; ok {
		topologySize.Resource = v.(string)
	}

	if v, ok := topologySizeMap["value"]; ok {
		topologySize.Value = int32(v.(int))
	}
}

func flattenElasticsearchClusterPlan(clusterInfo ElasticsearchClusterInfo) []map[string]interface{} {
	clusterPlanMaps := make([]map[string]interface{}, 1)

	clusterPlan := clusterInfo.PlanInfo.Current.Plan

	clusterPlanMap := make(map[string]interface{})
	clusterPlanMap["cluster_topology"] = flattenElasticsearchClusterTopology(clusterInfo, clusterPlan)
	clusterPlanMap["elasticsearch"] = flattenElasticsearchConfiguration(clusterPlan.Elasticsearch)

	clusterPlanMaps[0] = clusterPlanMap

	return clusterPlanMaps
}

func flattenElasticsearchClusterTopology(clusterInfo ElasticsearchClusterInfo, clusterPlan ElasticsearchClusterPlan) []map[string]interface{} {
	topologyMap := make([]map[string]interface{}, 0)

	// NOTE: This property appears as deprecated in the ECE API documentation, recommending use of the zone count from the
	// ElasticsearchClusterTopologyElement instead. However, zone count is not returned for ElasticsearchClusterTopologyElement
	// in the current version of ECE (2.2.3). To support either location, the zone count is used from cluster plan unless the
	// cluster topology element has a non-zero value.
	// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ElasticsearchClusterPlan
	defaultZoneCount := 1

	for i, t := range clusterPlan.ClusterTopology {
		elementMap := make(map[string]interface{})

		elementMap["instance_configuration_id"] = t.InstanceConfigurationID
		//elementMap["memory_per_node"] = t.MemoryPerNode
		//elementMap["node_count_per_zone"] = t.NodeCountPerZone

		elementMap["node_type"] = flattenElasticsearchNodeType(clusterInfo, i)

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
	//elasticsearchMap["system_settings"] = flattenElasticsearchSystemSettings(configuration.SystemSettings)

	elasticsearchMaps[0] = elasticsearchMap

	logJSON("Flattened elasticsearch configuration", elasticsearchMaps)

	return elasticsearchMaps
}

func flattenElasticsearchNodeType(clusterInfo ElasticsearchClusterInfo, instanceIndex int) map[string]interface{} {
	nodeTypeMap := make(map[string]interface{})

	if len(clusterInfo.Topology.Instances) > 0 {
		instance := clusterInfo.Topology.Instances[instanceIndex]

		nodeType := &ElasticsearchNodeType{}

		if instance.ServiceRoles != nil {
			nodeTypeValues := make(map[string]interface{})
			for _, s := range instance.ServiceRoles {
				nodeTypeValues[s] = true
			}

			expandElasticsearchNodeTypeFromMap(nodeType, nodeTypeValues)
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

func flattenElasticsearchSystemSettings(systemSettings ElasticsearchSystemSettings) []map[string]interface{} {
	systemSettingsMaps := make([]map[string]interface{}, 1)

	systemSettingsMap := make(map[string]interface{})
	systemSettingsMap["use_disk_threshold"] = systemSettings.UseDiskThreshold

	systemSettingsMaps[0] = systemSettingsMap

	logJSON("Flattened elasticsearch system settings", systemSettingsMaps)

	return systemSettingsMaps
}

func logJSON(context string, m interface{}) {
	jsonBytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		log.Printf("[DEBUG] %s: error marshalling value as JSON: %s. %v", context, err, m)
	}

	log.Printf("[DEBUG] %s: %s", context, string(jsonBytes))
}

// func updateKibanaCluster(client *ECEClient, clusterID string, d *schema.ResourceData, meta interface{}) error {
// 	// Use the Kibana Cluster ID to determine if an existing cluster is being updated/removed
// 	// or a new cluster should be created.
// 	var kibanaClusterID string
// 	v, ok := d.GetOk("kibana_cluster_id")
// 	if ok {
// 		kibanaClusterID = v.(string)
// 	}

// 	// Create a KibanaCreateRequest from the resource inputs.
// 	kibanaRequest, err := expandKibanaCreateRequest(d, meta)
// 	if err != nil {
// 		return err
// 	}

// 	// If the Kibana cluster ID is empty, Terraform does not know of an existing Kibana cluster.
// 	// In this case, if a cluster create request was created from resource inputs, use that
// 	// request to create a new Kibana cluster.
// 	if kibanaClusterID == "" {
// 		if kibanaRequest != nil {
// 			// Associate the new Kibana cluster with the elasticsearch cluster.
// 			kibanaRequest.ElasticsearchClusterID = clusterID

// 			// Create a new Kibana cluster.
// 			kibanaResponse, err := client.CreateKibanaCluster(*kibanaRequest)
// 			if err != nil {
// 				return err
// 			}

// 			kibanaClusterID = kibanaResponse.KibanaClusterID
// 			log.Printf("[DEBUG] Created Kibana cluster ID: %s\n", kibanaClusterID)
// 		}
// 	} else {
// 		// If the Kibana cluster ID is not empty and a Kibana create request was constructed from
// 		// resource inputs, update the existing cluster.
// 		if kibanaRequest != nil {
// 			// Update the existing Kibana cluster name.
// 			metadata := ClusterMetadataSettings{
// 				ClusterName: kibanaRequest.ClusterName,
// 			}

// 			_, err = client.UpdateKibanaClusterMetadata(kibanaClusterID, metadata)
// 			if err != nil {
// 				return err
// 			}

// 			// Update the existing Kibana cluster.
// 			_, err = client.UpdateKibanaCluster(kibanaClusterID, kibanaRequest.Plan)
// 			if err != nil {
// 				return err
// 			}
// 		} else {
// 			// If the Kibana create request is nil but the Kibana cluster ID is not empty, the existing
// 			// Kibana cluster should be deleted.
// 			_, err = client.DeleteKibanaCluster(kibanaClusterID)
// 			if err != nil {
// 				return err
// 			}

// 			kibanaClusterID = ""
// 			d.Set("kibana_cluster_id", nil)
// 		}
// 	}

// 	// If a Kibana cluster was created or updated, wait for the operation to complete and
// 	// check for success of the plan activity.
// 	if kibanaClusterID != "" {
// 		// Wait for the cluster plan to be initiated.
// 		duration := time.Duration(5) * time.Second // 5 seconds
// 		time.Sleep(duration)

// 		err = client.WaitForKibanaClusterStatus(kibanaClusterID, "started", false)
// 		if err != nil {
// 			return err
// 		}

// 		// Confirm that the Kibana update plan was successfully applied.
// 		err = validateKibanaClusterPlanActivity(client, kibanaClusterID)
// 		if err != nil {
// 			return err
// 		}

// 		d.Set("kibana_cluster_id", kibanaClusterID)
// 	}

// 	return nil
// }

func validateElasticsearchClusterPlanActivity(client *ECEClient, clusterID string) error {
	resp, err := client.GetElasticsearchClusterPlanActivity(clusterID)
	if err != nil {
		return err
	}

	if resp.StatusCode == 404 {
		return fmt.Errorf("%q: elasticsearch cluster ID was not found after update", clusterID)
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

		return fmt.Errorf("%q: elasticsearch cluster update failed: %v", clusterID, logMessages)
	}

	return nil
}

func validateKibanaClusterPlanActivity(client *ECEClient, clusterID string) error {
	resp, err := client.GetKibanaClusterPlanActivity(clusterID)
	if err != nil {
		return err
	}

	if resp.StatusCode == 404 {
		return fmt.Errorf("%q: kibana cluster ID was not found after update", clusterID)
	}

	var clusterPlansInfo KibanaClusterPlansInfo
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

		return fmt.Errorf("%q: kibana cluster update failed: %v", clusterID, logMessages)
	}

	return nil
}
