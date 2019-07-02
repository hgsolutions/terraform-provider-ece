provider "ece" {
  url      = "http://ec2-107-22-141-112.compute-1.amazonaws.com:12400"
  username = "admin"
  password = "bKbLcp8uQ6SJyfpYRjIhmjijbwbRWQa6c8ntx2Cqu7u"
  insecure = true                                                      # to bypass certificate check
  timeout  = 600                                                       # timeout after 10 minutes
}

resource "ece_cluster" "test_cluster" {
  name                  = "Test Cluster 1"
  elasticsearch_version = "7.2.0"
  memory_per_node       = 2048
  node_count_per_zone   = 1

  node_type {
    data   = true
    ingest = true
    master = true
    ml     = false
  }

  zone_count = 1
}
