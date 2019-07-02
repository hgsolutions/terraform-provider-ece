provider "ece" {
  url      = "http://ec2-107-22-141-112.compute-1.amazonaws.com:12400"
  username = "admin"
  password = "bKbLcp8uQ6SJyfpYRjIhmjijbwbRWQa6c8ntx2Cqu7u"
  insecure = true                                                      # to bypass certificate check
  timeout  = 600                                                       # timeout after 10 minutes
}

resource "ece_cluster" "test_cluster" {
  name                  = "Test Cluster 42"
  elasticsearch_version = "7.2.0"
  memory_per_node       = 1024
  node_count_per_zone   = 1

  node_type {
    data   = true
    master = true
    ingest = true
    ml     = false
  }

  zone_count = 1
}

output "test_cluster_id" {
  value       = "${ece_cluster.test_cluster.id}"
  description = "The ID of the cluster cluster"
}

output "test_cluster_name" {
  value       = "${ece_cluster.test_cluster.name}"
  description = "The name of the cluster"
}

output "test_cluster_elasticsearch_version" {
  value       = "${ece_cluster.test_cluster.elasticsearch_version}"
  description = "The elasticsearch version for the cluster"
}

output "test_cluster_node_type" {
  value       = "${ece_cluster.test_cluster.node_type}"
  description = "The node type for the cluster"
}

output "test_cluster_username" {
  value       = "${ece_cluster.test_cluster.elasticsearch_username}"
  description = "The username for logging in to the cluster."
}

output "test_cluster_password" {
  value       = "${ece_cluster.test_cluster.elasticsearch_password}"
  description = "The password for logging in to the cluster."
  sensitive   = false
}