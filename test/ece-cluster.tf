provider "ece" {
  url      = "http://ec2-18-234-124-116.compute-1.amazonaws.com:12400"
  username = "admin"
  password = "5tWLarHsRDWI7pAHHA1hqMwrhHqp0QhlaK70NizfRHl"
  insecure = true                                                      # to bypass certificate check
  timeout  = 600                                                       # timeout after 10 minutes
}

resource "ece_cluster" "test_cluster" {
  name                  = "My Test Cluster 3"
  elasticsearch_version = "7.1.0"
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
