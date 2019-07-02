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
	ElasticsearchClusterID string             `json:"elasticsearch_cluster_id"`
	Credentials            ClusterCredentials `json:"credentials"`
}

// ClusterMetadataSettings defines the top-level configuration settings for the Elasticsearch cluster.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ClusterMetadataSettings
type ClusterMetadataSettings struct {
	ClusterName string `json:"name"`
}

// CreateElasticsearchClusterRequest defines the request body for creating an Elasticsearch cluster.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#CreateElasticsearchClusterRequest
type CreateElasticsearchClusterRequest struct {
	ClusterName string                   `json:"cluster_name"`
	Plan        ElasticsearchClusterPlan `json:"plan"`
}

// ElasticsearchClusterInfo defines the information for an Elasticsearch cluster.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ElasticsearchClusterInfo
type ElasticsearchClusterInfo struct {
	ClusterID   string `json:"cluster_id"`
	ClusterName string `json:"cluster_name"`
	Healthy     bool   `json:"healthy"`
	Status      string `json:"status"`
}

// ElasticsearchClusterPlan defines the plan for an Elasticsearch cluster.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ElasticsearchClusterPlan
type ElasticsearchClusterPlan struct {
	Elasticsearch   ElasticsearchConfiguration            `json:"elasticsearch"`
	ClusterTopology []ElasticsearchClusterTopologyElement `json:"cluster_topology"`
}

// ElasticsearchClusterPlanInfo defines information about an Elasticsearch cluster plan.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ElasticsearchClusterPlanInfo
type ElasticsearchClusterPlanInfo struct {
	Healthy bool                     `json:"healthy"`
	Plan    ElasticsearchClusterPlan `json:"plan"`
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
