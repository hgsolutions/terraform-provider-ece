# terraform-provider-ece

Terraform provider for provisioning Elastic Cloud Enterprise (ECE) Elasticsearch clusters, compatible with v2.2 of ECE. 

Based on work by Phillip Baker: [terraform-provider-elasticsearch](https://github.com/phillbaker/terraform-provider-elasticsearch).

## Installation

Build or download a binary from the releases folder and put it in your Terraform user plugins directory.

See [the docs for more information](https://www.terraform.io/docs/plugins/basics.html).

## Usage

### Usage Notes

- Only a subset of the ECE API configuration parameters are currently implemented. See the `CreateElasticsearchClusterRequest` structure in the `ece_api_structures.go` file for the currently supported parameters.

- Configuration changes to existing clusters are applied using a cluster plan. This plan is evaluated by ECE to determine what changes are required to the existing cluster. Plans typically result in provisioning of new nodes and decommissioning of existing nodes.

- Not all combination of configuration parameters are supported by all ECE editions. For example, the open-source edition of ECE does not support Machine Learning (ML) nodes. If an unsupported configuration is specified, the ECE REST API may respond immediately with an error message, or the cluster plan may fail. In either case, the provider will respond with the ECE error message and indicate that the create or update failed.

### Sample Provider and Cluster Terraform configuration

```tf
provider "ece" {
  url      = "http://ece-api-url:12400"
  username = "admin"
  password = "************"
  insecure = true     # to bypass certificate checks
  timeout  = 600      # timeout after 10 minutes
}

resource "ece_cluster" "test_cluster" {
  cluster_name = "Test Cluster 42"

  plan {
    elasticsearch {
      version = "7.2.0"
    }

    cluster_topology {
      memory_per_node = 1024

      node_type {
        master = true
        data   = false
        ingest = true
      }
    }
  }
}
```

### Examples

#### Create a default cluster
To create a cluster with only the required inputs, use a configuration like the following. The created cluster will have a default topology of a single 1GB node instance with master, data, and ingest roles.

**NOTE:** The elasticsearch version is required and must be one of the Elastic Stack versions available in your ECE environment.

```
resource "ece_cluster" "test_cluster" {
  cluster_name = "Test Cluster 1"

  plan {
    elasticsearch {
      version = "7.2.0"
    }
  }
}
```

#### Create a cluster with separate master and data nodes
To create a cluster with separate master and data nodes, use a configuration like the following. The example also shows how to obtain outputs from each of the two topology elements.

```
resource "ece_cluster" "test_cluster" {
  cluster_name = "Test Cluster 2"

  plan {
    elasticsearch {
      version = "7.2.0"
    }

    cluster_topology {
      node_type {
        master = true
        data   = false
        ingest = true
      }
    }

    cluster_topology {
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

output "test_cluster_plan" {
  value       = "${ece_cluster.test_cluster.plan}"
  description = "The provided input plan for the cluster"
}

output "test_cluster_topology_0_node_count_per_zone" {
  value       = "${ece_cluster.test_cluster.plan.0.cluster_topology.0.node_count_per_zone}"
  description = "The node count per zone of the first topology element in the cluster"
}

output "test_cluster_topology_0_node_type_master" {
  value       = "${ece_cluster.test_cluster.plan.0.cluster_topology.0.node_type.0.master}"
  description = "Whether the role for the the first topology element in the cluster includes master"
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
```

## Development

### Requirements

* [Golang](https://golang.org/dl/) >= 1.7
* [Glide](https://github.com/Masterminds/glide)
* [ECE](https://www.elastic.co/downloads/enterprise)

### ECE Setup

To create a test ECE environment in AWS, the following will get you started:

1) Create a new AWS security group with the correct [port access for ECE](https://www.elastic.co/guide/en/cloud-enterprise/current/ece-prereqs-networking.html).

2) Launch a new EC2 instance from an `elastic-cloud-enterprise` Community AMI, specifying the above security group.

    * The ECE Ubuntu AMIs have most of the prereq configuration done on them for ECE, unlike Centos. For example, the `elastic-cloud-enterprise-xenial-201903110432` AMI is a good starting point.

    * Chose an instance type with the [minimum required hardware for ECE](https://www.elastic.co/guide/en/cloud-enterprise/current/ece-prereqs-hardware.html). For example, `r5.xlarge` could be used for a dev cluster.

2) Go through the [prerequisite configuration](https://www.elastic.co/guide/en/cloud-enterprise/current/ece-prereqs-software.html) for your chosen OS.

3) Download and run the installation script per the instructions here: https://www.elastic.co/guide/en/cloud-enterprise/current/ece-installing-online.html#ece-installing-first

### Debugging

By default, provider log messages are not written to standard out during provider execution. To enable verbose output of Terraform and provider log messages, set the `TF_LOG` environment variable to `DEBUG`.

### Building

#### For building on macOS

Ensure that this folder is at the following location: `${GOPATH}/src/github.com/Ascendon/terraform-provider-ece`

```
cd ~/go/src/github.com/Ascendon/terraform-provider-ece

glide install

go build -o releases/terraform-provider-ece_v0.2.1

cp releases/terraform-provider-ece_v0.2.1 ~/.terraform.d/plugins/darwin_amd64/.
```

## Contributing

1. Fork it ( https://github.com/Ascendon/terraform-provider-ece/fork )
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create a new Pull Request
