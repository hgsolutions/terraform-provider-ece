provider "ece" {
  url      = "http://ec2-123-456-789-101.compute-1.amazonaws.com:12400"
  username = "admin"
  password = "******"
  insecure = true                                                      # to bypass certificate check
  timeout  = 600                                                       # timeout after 10 minutes
}

resource "ece_cluster" "test_cluster" {
  cluster_name = "Test Cluster 42"

  // See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ElasticsearchClusterPlan
  plan {
    // See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ElasticsearchConfiguration
    elasticsearch {
      version = "7.2.0"
    }

    // See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ElasticsearchClusterTopologyElement
    cluster_topology {
      memory_per_node = 1024

      // See https://www.elastic.co/guide/en/cloud-enterprise/current/definitions.html#ElasticsearchNodeType
      node_type {
        master = true
        data   = false
        ingest = true
      }
    }

    cluster_topology {
      memory_per_node = 1024

      node_type {
        master = false
        data   = true
        ingest = true
      }
    }
  }
}

output "test_cluster_id" {
  value       = "${ece_cluster.test_cluster.id}"
  description = "The ID of the cluster"
}

output "test_cluster_name" {
  value       = "${ece_cluster.test_cluster.cluster_name}"
  description = "The name of the cluster"
}

output "test_cluster_elasticsearch_version" {
  value       = "${ece_cluster.test_cluster.plan.0.elasticsearch.0.version}"
  description = "The elasticsearch version of the cluster"
}

output "test_cluster_topology" {
  value       = "${ece_cluster.test_cluster.plan.0.cluster_topology}"
  description = "The topology of the cluster."
}

output "test_cluster_topology_0_node_count_per_zone" {
  value       = "${ece_cluster.test_cluster.plan.0.cluster_topology.0.node_count_per_zone}"
  description = "The node count per zone of first topology element in the cluster"
}

output "test_cluster_topology_1_memory_per_node" {
  value       = "${ece_cluster.test_cluster.plan.0.cluster_topology.1.memory_per_node}"
  description = "The memory per node for the second topology element in the cluster"
}

output "test_cluster_username" {
  value       = "${ece_cluster.test_cluster.elasticsearch_username}"
  description = "The username for logging in to the cluster."
}

output "test_cluster_password" {
  value       = "${ece_cluster.test_cluster.elasticsearch_password}"
  description = "The password for logging in to the cluster."
  sensitive   = true
}
