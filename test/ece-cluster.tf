provider "ece" {
  url      = "cloud.elastic.co"
  username = "admin"
  password = "******"
  insecure = false                                                     # to bypass certificate check
  timeout  = 600                                                       # timeout after 10 minutes
}

resource "ece_elasticsearch_cluster" "test_cluster" {
  cluster_name = "Test Cluster 42"

  plan {
    elasticsearch {
      version = "7.6.0"
    }

    cluster_topology {
      memory_per_node = 1024

      node_type {
        master = true
        data   = false
        ingest = true
      }
    }

    cluster_topology {
      memory_per_node = 2048

      node_type {
        master = false
        data   = true
        ingest = true
      }
    }
  }

  kibana {
    cluster_name = "Test Cluster 42"

    plan {
      cluster_topology {
        memory_per_node = 2048
        node_count_per_zone = 2
        zone_count = 1
      }
    }
  }
}

output "test_cluster_id" {
  value       = "${ece_elasticsearch_cluster.test_cluster.id}"
  description = "The ID of the cluster"
}

output "test_cluster_name" {
  value       = "${ece_elasticsearch_cluster.test_cluster.cluster_name}"
  description = "The name of the cluster"
}

output "test_cluster_elasticsearch_version" {
  value       = "${ece_elasticsearch_cluster.test_cluster.plan.0.elasticsearch.0.version}"
  description = "The elasticsearch version of the cluster"
}

output "test_cluster_topology" {
  value       = "${ece_elasticsearch_cluster.test_cluster.plan.0.cluster_topology}"
  description = "The topology of the cluster."
}

output "test_cluster_username" {
  value       = "${ece_elasticsearch_cluster.test_cluster.elasticsearch_username}"
  description = "The username for logging in to the cluster."
}

output "test_cluster_password" {
  value       = "${ece_elasticsearch_cluster.test_cluster.elasticsearch_password}"
  description = "The password for logging in to the cluster."
  sensitive   = true
}
