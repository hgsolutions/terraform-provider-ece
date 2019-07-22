package main

// ClusterCredentials defines the username and password for the new Elasticsearch cluster, which
// is returned from the Elasticsearch cluster create command.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ClusterCredentials
type ClusterCredentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

// ClusterCrudResponse defines the response to an Elasticsearch cluster or Kibana instance CRUD
// (create/update-plan) request.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ClusterCrudResponse
type ClusterCrudResponse struct {
	Credentials            ClusterCredentials `json:"credentials"`
	ElasticsearchClusterID string             `json:"elasticsearch_cluster_id"`
	KibanaClusterID        string             `json:"kibana_cluster_id"`
}

// ClusterInstanceInfo defines information about each instance in the Elasticsearch cluster.
type ClusterInstanceInfo struct {
	ServiceRoles []string `json:"service_roles"` // Currently only populated for Elasticsearch, with possible values: master,data,ingest,ml
}

// ClusterMetadataSettings defines the top-level configuration settings for the Elasticsearch cluster.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ClusterMetadataSettings
type ClusterMetadataSettings struct {
	ClusterName string `json:"name"`
}

// ClusterPlanStepInfo defines information about a step in a plan.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ClusterPlanStepInfo
type ClusterPlanStepInfo struct {
	Completed  string                          `json:"completed"`
	DurationMS int64                           `json:"duration_in_millis"`
	InfoLog    []ClusterPlanStepLogMessageInfo `json:"info_log"`
	Stage      string                          `json:"stage"`
	Started    string                          `json:"started"`
	Status     string                          `json:"status"`
	StepID     string                          `json:"step_id"`
}

// ClusterPlanStepLogMessageInfo defines the log message from a specified stage of an executed step in a plan.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ClusterPlanStepLogMessageInfo
type ClusterPlanStepLogMessageInfo struct {
	DeltaMS   int64  `json:"delta_in_millis"`
	Message   string `json:"message"`
	Stage     string `json:"stage"`
	Timestamp string `json:"timestamp"`
}

// ClusterTopologyInfo defines the topology for Elasticsearch clusters, multiple Kibana instances, or multiple APM Servers.
// The ClusterTopologyInfo also includes the instances and containers, and where they are located.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ClusterTopologyInfo
type ClusterTopologyInfo struct {
	Healthy   bool                  `json:"healthy"`
	Instances []ClusterInstanceInfo `json:"instances"`
}

// CreateElasticsearchClusterRequest defines the request body for creating an Elasticsearch cluster.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#CreateElasticsearchClusterRequest
type CreateElasticsearchClusterRequest struct {
	ClusterName string                                    `json:"cluster_name"`
	Kibana      *CreateKibanaInCreateElasticsearchRequest `json:"kibana"`
	Plan        ElasticsearchClusterPlan                  `json:"plan"`
}

// CreateKibanaInCreateElasticsearchRequest defines the request body for creating a Kibana instance,
// which is included in the Elasticsearch cluster create request.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#CreateKibanaInCreateElasticsearchRequest
type CreateKibanaInCreateElasticsearchRequest struct {
	ClusterName string             `json:"cluster_name,omitempty"`
	Plan        *KibanaClusterPlan `json:"plan"`
}

// CreateKibanaRequest defines the request body for creating one or more Kibana instances.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#CreateKibanaRequest
type CreateKibanaRequest struct {
	ClusterName            string             `json:"cluster_name"`
	ElasticsearchClusterID string             `json:"elasticsearch_cluster_id"`
	Plan                   *KibanaClusterPlan `json:"plan"`
}

// ElasticsearchClusterInfo defines the information for an Elasticsearch cluster.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ElasticsearchClusterInfo
type ElasticsearchClusterInfo struct {
	ClusterID                string                        `json:"cluster_id"`
	ClusterName              string                        `json:"cluster_name"`
	Healthy                  bool                          `json:"healthy"`
	PlanInfo                 ElasticsearchClusterPlansInfo `json:"plan_info"`
	AssociatedKibanaClusters []KibanaSubClusterInfo        `json:"associated_kibana_clusters"`
	Status                   string                        `json:"status"`
	Topology                 ClusterTopologyInfo           `json:"topology"`
}

// ElasticsearchClusterPlan defines the plan for an Elasticsearch cluster.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ElasticsearchClusterPlan
type ElasticsearchClusterPlan struct {
	ClusterTopology []ElasticsearchClusterTopologyElement   `json:"cluster_topology"`
	Elasticsearch   ElasticsearchConfiguration              `json:"elasticsearch"`
	Transient       TransientElasticsearchPlanConfiguration `json:"transient,omitempty"`
	ZoneCount       int                                     `json:"zone_count"`
}

// ElasticsearchClusterPlanInfo defines information about an Elasticsearch cluster plan.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ElasticsearchClusterPlanInfo
type ElasticsearchClusterPlanInfo struct {
	AttemptEndTime   string                   `json:"attempt_end_time"`
	AttemptStartTime string                   `json:"attempt_start_time"`
	Healthy          bool                     `json:"healthy"`
	Plan             ElasticsearchClusterPlan `json:"plan"`
	PlanAttemptID    string                   `json:"plan_attempt_id"`
	PlanAttemptLog   []ClusterPlanStepInfo    `json:"plan_attempt_log"`
	PlanAttemptName  string                   `json:"plan_attempt_name"`
	PlanEndTime      string                   `json:"plan_end_time"`
}

// ElasticsearchClusterPlansInfo defines information about the current, pending, and past Elasticsearch cluster plans.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ElasticsearchClusterPlansInfo
type ElasticsearchClusterPlansInfo struct {
	Current ElasticsearchClusterPlanInfo   `json:"current"`
	Healthy bool                           `json:"healthy"`
	History []ElasticsearchClusterPlanInfo `json:"history"`
	Pending ElasticsearchClusterPlanInfo   `json:"pending"`
}

// ElasticsearchClusterTopologyElement defines the topology of the Elasticsearch nodes, including the number,
// capacity, and type of nodes, and where they can be allocated.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ElasticsearchClusterTopologyElement
type ElasticsearchClusterTopologyElement struct {
	MemoryPerNode    int                   `json:"memory_per_node"`
	NodeCountPerZone int                   `json:"node_count_per_zone"`
	NodeType         ElasticsearchNodeType `json:"node_type"`
	ZoneCount        int                   `json:"zone_count"`
}

// DefaultElasticsearchClusterTopologyElement returns a new ElasticsearchClusterTopologyElement with default values.
func DefaultElasticsearchClusterTopologyElement() *ElasticsearchClusterTopologyElement {
	return &ElasticsearchClusterTopologyElement{
		MemoryPerNode:    1024,
		NodeCountPerZone: 1,
		NodeType:         *DefaultElasticsearchNodeType(),
		ZoneCount:        1,
	}
}

// ElasticsearchConfiguration defines the Elasticsearch cluster settings.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ElasticsearchConfiguration
type ElasticsearchConfiguration struct {
	Version string `json:"version"`
}

// ElasticsearchNodeType defines the combinations of Elasticsearch node types.
// TIP: By default, the Elasticsearch node is master eligible, can hold data, and run ingest pipelines.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ElasticsearchNodeType
type ElasticsearchNodeType struct {
	Data   bool `json:"data"`
	Ingest bool `json:"ingest"`
	Master bool `json:"master"`
	ML     bool `json:"ml"`
}

// DefaultElasticsearchNodeType creates a new ElasticsearchNodeType with default values.
func DefaultElasticsearchNodeType() *ElasticsearchNodeType {
	return &ElasticsearchNodeType{
		Data:   true,
		Ingest: true,
		Master: true,
		ML:     false,
	}
}

// ElasticsearchPlanControlConfiguration defines the configuration settings for the timeout and fallback parameters.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ElasticsearchPlanControlConfiguration
type ElasticsearchPlanControlConfiguration struct {
	// Commenting because default is calculated based on cluster size and is
	// typically higher than configured provider timeout.
	// Timeout int64 `json:"timeout"`
}

// KibanaClusterInfo defines the top-level object information for a Kibana instance.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#KibanaClusterInfo
type KibanaClusterInfo struct {
	ClusterID   string `json:"cluster_id"`
	ClusterName string `json:"cluster_name"`
	Healthy     bool   `json:"healthy"`
	Status      string `json:"status"`
}

// KibanaClusterPlan defines the plan for the Kibana instance.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#KibanaClusterPlan
type KibanaClusterPlan struct {
	ClusterTopology []KibanaClusterTopologyElement `json:"cluster_topology"`
	Kibana          KibanaConfiguration            `json:"kibana"`
	ZoneCount       int                            `json:"zone_count"`
}

// KibanaClusterPlanInfo defines information about the current, pending, or past Kibana instance plan.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#KibanaClusterPlanInfo
type KibanaClusterPlanInfo struct {
	AttemptEndTime   string                `json:"attempt_end_time"`
	AttemptStartTime string                `json:"attempt_start_time"`
	Healthy          bool                  `json:"healthy"`
	Plan             KibanaClusterPlan     `json:"plan"`
	PlanAttemptID    string                `json:"plan_attempt_id"`
	PlanAttemptLog   []ClusterPlanStepInfo `json:"plan_attempt_log"`
	PlanAttemptName  string                `json:"plan_attempt_name"`
	PlanEndTime      string                `json:"plan_end_time"`
}

// KibanaClusterPlansInfo defines information about the current, pending, or past Kibana instance plans.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#KibanaClusterPlansInfo
type KibanaClusterPlansInfo struct {
	Current KibanaClusterPlanInfo `json:"current"`
	Healthy bool                  `json:"healthy"`
}

// DefaultKibanaClusterPlan returns a new KibanaClusterPlan with default values.
func DefaultKibanaClusterPlan() *KibanaClusterPlan {
	return &KibanaClusterPlan{
		ClusterTopology: []KibanaClusterTopologyElement{*DefaultKibanaClusterTopologyElement()},
		Kibana:          *DefaultKibanaConfiguration(),
		ZoneCount:       1,
	}
}

// KibanaClusterTopologyElement defines the topology of the Kibana nodes, including the number, capacity, and
// type of nodes, and where they can be allocated.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#KibanaClusterTopologyElement
type KibanaClusterTopologyElement struct {
	MemoryPerNode int   `json:"memory_per_node"`
	ZoneCount     int32 `json:"zone_count"`
}

// DefaultKibanaClusterTopologyElement returns a new KibanaClusterTopologyElement with default values.
func DefaultKibanaClusterTopologyElement() *KibanaClusterTopologyElement {
	return &KibanaClusterTopologyElement{
		MemoryPerNode: 1024,
		ZoneCount:     1,
	}
}

// KibanaConfiguration defines the Kibana instance settings. When specified at the top level, provides a field-by-field default.
// When specified at the topology level, provides the override settings.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#KibanaConfiguration
type KibanaConfiguration struct {
}

// DefaultKibanaConfiguration returns a new KibanaConfiguration with default values.
func DefaultKibanaConfiguration() *KibanaConfiguration {
	return &KibanaConfiguration{}
}

// KibanaSubClusterInfo defines information about the Kibana instances associated with the Elasticsearch cluster.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#KibanaSubClusterInfo
type KibanaSubClusterInfo struct {
	Enabled  bool   `json:"enabled"`
	KibanaID string `json:"kibana_id"`
}

// TransientElasticsearchPlanConfiguration defines the configuration parameters that control how the plan is applied.
// For example, the Elasticsearch cluster topology and Elasticsearch settings.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#TransientElasticsearchPlanConfiguration
type TransientElasticsearchPlanConfiguration struct {
	PlanConfiguration ElasticsearchPlanControlConfiguration `json:"plan_configuration"`
}
