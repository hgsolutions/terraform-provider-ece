package main

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

// CreateElasticsearchClusterRequest defines the request body for creating an Elasticsearch cluster.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#CreateElasticsearchClusterRequest
type CreateElasticsearchClusterRequest struct {
	ClusterName string                   `json:"cluster_name"`
	Plan        ElasticsearchClusterPlan `json:"plan"`
}

// ElasticsearchClusterPlan defines the plan for an Elasticsearch cluster.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ElasticsearchClusterPlan
type ElasticsearchClusterPlan struct {
	Elasticsearch   ElasticsearchConfiguration            `json:"elasticsearch"`
	ClusterTopology []ElasticsearchClusterTopologyElement `json:"cluster_topology"`
}

// ElasticsearchConfiguration defines the Elasticsearch cluster settings.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ElasticsearchConfiguration
type ElasticsearchConfiguration struct {
	Version string `json:"version"`
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

// ElasticsearchNodeType defines the combinations of Elasticsearch node types.
// TIP: By default, the Elasticsearch node is master eligible, can hold data, and run ingest pipelines.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ElasticsearchNodeType
type ElasticsearchNodeType struct {
	Data   bool `json:"data"`
	Ingest bool `json:"ingest"`
	Master bool `json:"master"`
	ML     bool `json:"ml"`
}

// ClusterCrudResponse defines the response to an Elasticsearch cluster or Kibana instance CRUD
// (create/update-plan) request.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ClusterCrudResponse
type ClusterCrudResponse struct {
	ElasticsearchClusterID string             `json:"elasticsearch_cluster_id"`
	Credentials            ClusterCredentials `json:"credentials"`
}

// ClusterCredentials defines the username and password for the new Elasticsearch cluster, which
// is returned from the Elasticsearch cluster create command.
// See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ClusterCredentials
type ClusterCredentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}
