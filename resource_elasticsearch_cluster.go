package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

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
		Create: resourceElasticsearchClusterCreate,
		Read:   resourceElasticsearchClusterRead,
		Update: resourceElasticsearchClusterUpdate,
		Delete: resourceElasticsearchClusterDelete,
		Schema: map[string]*schema.Schema{
			"cluster_name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The name of the cluster.",
				ForceNew:    false,
				Required:    true,
			},
			"kibana": {
				Type:        schema.TypeList,
				Description: "The plan for a Kibana instance that should be created as part of the Elasticsearch cluster.",
				ForceNew:    false,
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cluster_name": &schema.Schema{
							Type:        schema.TypeString,
							Description: "The name of the Kibana cluster.",
							ForceNew:    false,
							Optional:    true,
						},
						"plan": {
							Type:        schema.TypeList,
							Description: "The plan for the Kibana cluster.",
							ForceNew:    false,
							Optional:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cluster_topology": {
										Type:        schema.TypeList,
										Description: "The topology of the Kibana nodes, including the number, capacity, and type of nodes, and where they can be allocated.",
										Optional:    true,
										Computed:    false,
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"memory_per_node": &schema.Schema{
													Type:        schema.TypeInt,
													Description: "The memory capacity in MB for each node of this type built in each zone. The default is 1024.",
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
												"zone_count": &schema.Schema{
													Type:        schema.TypeInt,
													ForceNew:    false,
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
										Description: "The memory capacity in MB for each node of this type built in each zone. The default is 1024.",
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
			"kibana_cluster_id": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID for the created Kibana cluster.",
			},
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceElasticsearchClusterCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ECEClient)

	clusterName := d.Get("cluster_name").(string)
	log.Printf("[DEBUG] Creating elasticsearch cluster with name: %s\n", clusterName)

	clusterPlan, err := expandElasticsearchClusterPlan(d, meta)
	if err != nil {
		return err
	}

	createClusterRequest := CreateElasticsearchClusterRequest{
		ClusterName: clusterName,
		Plan:        *clusterPlan,
	}

	kibanaRequest, err := expandKibanaCreateRequest(d, meta)
	if err != nil {
		return err
	} else if kibanaRequest != nil {
		log.Printf("[DEBUG] Kibana instance will be created: %v\n", *kibanaRequest)
		createClusterRequest.Kibana = &CreateKibanaInCreateElasticsearchRequest{
			ClusterName: kibanaRequest.ClusterName,
			Plan:        kibanaRequest.Plan,
		}
	}

	crudResponse, err := client.CreateElasticsearchCluster(createClusterRequest)
	if err != nil {
		return err
	}

	elasticsearchClusterID := crudResponse.ElasticsearchClusterID
	log.Printf("[DEBUG] Created elasticsearch cluster ID: %s\n", elasticsearchClusterID)

	err = client.WaitForElasticsearchClusterStatus(elasticsearchClusterID, "started", false)
	if err != nil {
		return err
	}

	// Confirm that the elasticsearch creation plan was successfully applied.
	err = validateElasticsearchClusterPlanActivity(client, elasticsearchClusterID)
	if err != nil {
		return err
	}

	d.SetId(elasticsearchClusterID)
	d.Set("elasticsearch_username", crudResponse.Credentials.Username)
	d.Set("elasticsearch_password", crudResponse.Credentials.Password)

	// Wait for the Kibana cluster to be created if it was included in the creation request.
	kibanaClusterID := crudResponse.KibanaClusterID
	if kibanaClusterID != "" {
		err = client.WaitForKibanaClusterStatus(kibanaClusterID, "started", false)
		if err != nil {
			return err
		}

		// Confirm that the Kibana creation plan was successfully applied.
		err = validateKibanaClusterPlanActivity(client, kibanaClusterID)
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] Created Kibana cluster ID: %s\n", kibanaClusterID)
		d.Set("kibana_cluster_id", kibanaClusterID)
	}

	return resourceElasticsearchClusterRead(d, meta)
}

func resourceElasticsearchClusterRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ECEClient)

	clusterID := d.Id()
	log.Printf("[DEBUG] Reading elasticsearch cluster information for cluster ID: %s\n", clusterID)

	resp, err := client.GetElasticsearchCluster(clusterID)
	if err != nil {
		return err
	}

	// If the resource does not exist, inform Terraform. We want to immediately
	// return here to prevent further processing.
	if resp.StatusCode == 404 {
		log.Printf("[DEBUG] Elasticsearch cluster ID not found: %s\n", clusterID)
		d.SetId("")
		return nil
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Elasticsearch cluster response body: %v\n", string(respBytes))

	var clusterInfo ElasticsearchClusterInfo
	err = json.Unmarshal(respBytes, &clusterInfo)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Setting elasticsearch cluster_name: %v\n", clusterInfo.ClusterName)
	d.Set("cluster_name", clusterInfo.ClusterName)

	plan := flattenElasticsearchClusterPlan(clusterInfo)
	log.Printf("[DEBUG] Setting elasticsearch cluster plan: %v\n", plan)
	d.Set("plan", plan)
	if err != nil {
		return err
	}

	if clusterInfo.AssociatedKibanaClusters != nil && len(clusterInfo.AssociatedKibanaClusters) > 0 {
		kibanaClusterID := clusterInfo.AssociatedKibanaClusters[0].KibanaID
		log.Printf("[DEBUG] Setting Kibana cluster ID: %v\n", kibanaClusterID)
		d.Set("kibana_cluster_id", kibanaClusterID)
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceElasticsearchClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ECEClient)

	d.Partial(true)

	clusterID := d.Id()
	log.Printf("[DEBUG] Updating elasticsearch cluster ID: %s\n", clusterID)

	resp, err := client.GetElasticsearchCluster(clusterID)
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

		_, err = client.UpdateElasticsearchClusterMetadata(clusterID, metadata)
		if err != nil {
			return err
		}
	}

	d.SetPartial("cluster_name")

	if d.HasChange("plan") {
		clusterPlan, err := expandElasticsearchClusterPlan(d, meta)
		if err != nil {
			return err
		}

		_, err = client.UpdateElasticsearchCluster(clusterID, *clusterPlan)
		if err != nil {
			return err
		}

		// Wait for the cluster plan to be initiated.
		duration := time.Duration(5) * time.Second // 5 seconds
		time.Sleep(duration)

		err = client.WaitForElasticsearchClusterStatus(clusterID, "started", false)
		if err != nil {
			return err
		}

		// Confirm that the update plan was successfully applied.
		err = validateElasticsearchClusterPlanActivity(client, clusterID)
		if err != nil {
			return err
		}
	}

	d.SetPartial("plan")

	if d.HasChange("kibana") {
		err = updateKibanaCluster(client, clusterID, d, meta)
		if err != nil {
			return err
		}
	}

	d.Partial(false)

	return resourceElasticsearchClusterRead(d, meta)
}

func resourceElasticsearchClusterDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ECEClient)
	clusterID := d.Id()

	log.Printf("[DEBUG] Deleting cluster ID: %s\n", clusterID)
	_, err := client.DeleteElasticsearchCluster(clusterID)
	if err != nil {
		return err
	}

	return nil
}

func expandElasticsearchClusterPlan(d *schema.ResourceData, meta interface{}) (clusterPlan *ElasticsearchClusterPlan, err error) {
	clusterPlanList := d.Get("plan").([]interface{})
	clusterPlanMap := clusterPlanList[0].(map[string]interface{})

	clusterTopology := expandElasticsearchClusterTopology(clusterPlanMap)
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

func expandElasticsearchClusterTopology(clusterPlanMap map[string]interface{}) []ElasticsearchClusterTopologyElement {
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
				expandElasticsearchNodeTypeFromMap(nodeType, nodeTypeMaps[0].(map[string]interface{}))
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

	if v, ok := kibanaPlanMap["zone_count"]; ok {
		kibanaPlan.ZoneCount = v.(int)
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
		if v, ok := elementMap["memory_per_node"]; ok {
			clusterTopologyElement.MemoryPerNode = v.(int)
		}

		if v, ok := elementMap["node_count_per_zone"]; ok {
			clusterTopologyElement.NodeCountPerZone = v.(int)
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

func expandKibanaCreateRequest(d *schema.ResourceData, meta interface{}) (kibanaRequest *CreateKibanaRequest, err error) {
	kibanaList := d.Get("kibana").([]interface{})

	if kibanaList == nil || len(kibanaList) == 0 {
		log.Printf("[DEBUG] Kibana configuration not specified. No Kibana instance will be created.\n")
		return nil, nil
	}

	kibanaPlan := DefaultKibanaClusterPlan()

	if kibanaList[0] == nil {
		log.Printf("[DEBUG] Empty Kibana configuration specified. A default Kibana instance will be created.\n")

		kibanaRequest = &CreateKibanaRequest{
			Plan: kibanaPlan,
		}

		return kibanaRequest, nil
	}

	var kibanaName string
	kibanaMap := kibanaList[0].(map[string]interface{})

	if v, ok := kibanaMap["cluster_name"]; ok {
		kibanaName = v.(string)
	}

	if v, ok := kibanaMap["plan"]; ok {
		err := expandKibanaClusterPlan(kibanaPlan, v.(interface{}))
		if err != nil {
			return nil, err
		}
	}

	kibanaRequest = &CreateKibanaRequest{
		ClusterName: kibanaName,
		Plan:        kibanaPlan,
	}

	return kibanaRequest, nil
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
	defaultZoneCount := clusterPlan.ZoneCount

	for i, t := range clusterPlan.ClusterTopology {
		elementMap := make(map[string]interface{})

		elementMap["memory_per_node"] = t.MemoryPerNode
		elementMap["node_count_per_zone"] = t.NodeCountPerZone

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

func logJSON(context string, m interface{}) {
	jsonBytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		log.Printf("[DEBUG] %s: error marshalling value as JSON: %s. %v", context, err, m)
	}

	log.Printf("[DEBUG] %s: %s", context, string(jsonBytes))
}

func updateKibanaCluster(client *ECEClient, clusterID string, d *schema.ResourceData, meta interface{}) error {
	// Use the Kibana Cluster ID to determine if an existing cluster is being updated/removed
	// or a new cluster should be created.
	var kibanaClusterID string
	v, ok := d.GetOk("kibana_cluster_id")
	if ok {
		kibanaClusterID = v.(string)
	}

	// Create a KibanaCreateRequest from the resource inputs.
	kibanaRequest, err := expandKibanaCreateRequest(d, meta)
	if err != nil {
		return err
	}

	// If the Kibana cluster ID is empty, Terraform does not know of an existing Kibana cluster.
	// In this case, if a cluster create request was created from resource inputs, use that
	// request to create a new Kibana cluster.
	if kibanaClusterID == "" {
		if kibanaRequest != nil {
			// Associate the new Kibana cluster with the elasticsearch cluster.
			kibanaRequest.ElasticsearchClusterID = clusterID

			// Create a new Kibana cluster.
			kibanaResponse, err := client.CreateKibanaCluster(*kibanaRequest)
			if err != nil {
				return err
			}

			kibanaClusterID = kibanaResponse.KibanaClusterID
			log.Printf("[DEBUG] Created Kibana cluster ID: %s\n", kibanaClusterID)
		}
	} else {
		// If the Kibana cluster ID is not empty and a Kibana create request was constructed from
		// resource inputs, update the existing cluster.
		if kibanaRequest != nil {
			// Update the existing Kibana cluster name.
			metadata := ClusterMetadataSettings{
				ClusterName: kibanaRequest.ClusterName,
			}

			_, err = client.UpdateKibanaClusterMetadata(kibanaClusterID, metadata)
			if err != nil {
				return err
			}

			// Update the existing Kibana cluster.
			_, err = client.UpdateKibanaCluster(kibanaClusterID, kibanaRequest.Plan)
			if err != nil {
				return err
			}
		} else {
			// If the Kibana create request is nil but the Kibana cluster ID is not empty, the existing
			// Kibana cluster should be deleted.
			_, err = client.DeleteKibanaCluster(kibanaClusterID)
			if err != nil {
				return err
			}

			kibanaClusterID = ""
			d.Set("kibana_cluster_id", nil)
		}
	}

	// If a Kibana cluster was created or updated, wait for the operation to complete and
	// check for success of the plan activity.
	if kibanaClusterID != "" {
		// Wait for the cluster plan to be initiated.
		duration := time.Duration(5) * time.Second // 5 seconds
		time.Sleep(duration)

		err = client.WaitForKibanaClusterStatus(kibanaClusterID, "started", false)
		if err != nil {
			return err
		}

		// Confirm that the Kibana update plan was successfully applied.
		err = validateKibanaClusterPlanActivity(client, kibanaClusterID)
		if err != nil {
			return err
		}

		d.Set("kibana_cluster_id", kibanaClusterID)
	}

	return nil
}

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
