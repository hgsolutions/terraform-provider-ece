# terraform-provider-ece

[![Build Status](https://travis-ci.org/Ascendon/terraform-provider-ece.svg?branch=master)](https://travis-ci.org/Ascendon/terraform-provider-ece)

Terraform provider for provisioning Elastic Cloud Enterprise (ECE) Elasticsearch clusters, compatible with v2.2 of ECE. 

Based on work by Phillip Baker: [terraform-provider-elasticsearch](https://github.com/phillbaker/terraform-provider-elasticsearch).

## Installation

Build or download a release binary and place it in your Terraform user plugins directory.

See [the docs for more information](https://www.terraform.io/docs/plugins/basics.html).

## Usage

### Usage Notes

- The general structure of the `ece_elasticsearch_cluster` resource schema was designed to match the request structure for creation of new Elasticsearch clusters using the ECE REST API. This API is documented here: https://www.elastic.co/guide/en/cloud-enterprise/current/ece-api-reference.html

- Several of the aggregate schema resources would be better mapped as TypeMap, but currently TypeMap cannot be used for non-string values due to this bug: https://github.com/hashicorp/terraform/issues/15327. As a result, I used TypeList with a MaxValue of 1, matching what is done with the AWS provider for Elasticsearch domains. A consequence is that outputs of nested values will require an index designation, even when only one subitem is allowed. For example, to retrieve the node_count_per_zone from the cluster plan's first topology element, you would need to use this approach:

  ```
  ece_elasticsearch_cluster.test_cluster.plan.0.cluster_topology.0.node_count_per_zone
  ```

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

resource "ece_elasticsearch_cluster" "test_cluster" {
  cluster_name = "Test Cluster 42"

  plan {
    elasticsearch {
      version = "7.2.0"
    }

    cluster_topology {
      memory_per_node = 2048

      node_type {
        master = true
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
      }
    }
  }
}
```
### Resources
The provider currently supports a single resource: 

### `ece_elasticsearch_cluster`
This resource creates an ECE Elasticsearch cluster and, optionally, an associated Kibana cluster.

#### Resource Outputs
The following outputs are available after `ece_elasticsearch_cluster` resource creation:

- `id`: the ID for the created Elasticsearch cluster

- `elasticsearch_username`: the username for the created Elasticsearch cluster

- `elasticsearch_password`: the password for the created Elasticsearch cluster

- `kibana_cluster_id`: the ID for the created Kibana cluster, if any

#### Examples

#### Create a default Elasticsearch cluster
To create an Elasticsearch cluster with only the required inputs, use a configuration like the following. The created cluster will have a default topology of a single 1GB node instance with master, data, and ingest roles.

**NOTE:** The Elasticsearch version is required and must be one of the Elastic Stack versions available in your ECE environment.

```
resource "ece_elasticsearch_cluster" "test_cluster" {
  cluster_name = "Test Cluster 1"

  plan {
    elasticsearch {
      version = "7.2.0"
    }
  }
}

output "test_cluster_id" {
  value       = "${ece_elasticsearch_cluster.test_cluster.id}"
  description = "The ID of the cluster"
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
```

#### Create an Elasticsearch cluster with separate master and data nodes
To create an Elasticsearch cluster with separate master and data nodes, use a configuration like the following. The example also shows how to obtain outputs from each of the two topology elements.

```
resource "ece_elasticsearch_cluster" "test_cluster" {
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

output "test_cluster_plan" {
  value       = "${ece_elasticsearch_cluster.test_cluster.plan}"
  description = "The provided input plan for the cluster"
}

output "test_cluster_topology_0_node_count_per_zone" {
  value       = "${ece_elasticsearch_cluster.test_cluster.plan.0.cluster_topology.0.node_count_per_zone}"
  description = "The node count per zone of the first topology element in the cluster"
}

output "test_cluster_topology_0_node_type_master" {
  value       = "${ece_elasticsearch_cluster.test_cluster.plan.0.cluster_topology.0.node_type.0.master}"
  description = "Whether the role for the the first topology element in the cluster includes master"
}

output "test_cluster_topology_1_memory_per_node" {
  value       = "${ece_elasticsearch_cluster.test_cluster.plan.0.cluster_topology.1.memory_per_node}"
  description = "The memory per node for the second topology element in the cluster"
}
```

#### Create an Elasticsearch cluster with an associated default Kibana cluster.
To create an Elasticsearch cluster with an associatd default Kibana cluster, use a configuration like the following.

```
resource "ece_elasticsearch_cluster" "test_cluster" {
  cluster_name = "Test Cluster 3"

  plan {
    elasticsearch {
      version = "7.2.0"
    }
  }

  kibana {
  }
}

output "test_kibana_cluster_id" {
  value       = "${ece_elasticsearch_cluster.test_cluster.kibana_cluster_id}"
  description = "The ID of the Kibana cluster"
}
```

#### Create an Elasticsearch cluster with an associated configured Kibana cluster.
To create an Elasticsearch cluster with an associatd configured Kibana cluster, use a configuration like the following.

```
resource "ece_elasticsearch_cluster" "test_cluster" {
  cluster_name = "Test Cluster 4"

  plan {
    elasticsearch {
      version = "7.2.0"
    }
  }

  kibana {
    cluster_name = "Test Cluster 4"

    plan {
      cluster_topology {
        memory_per_node = 2048
        node_count_per_zone = 2
        zone_count = 1
      }
    }
  }
}

output "test_kibana_cluster_id" {
  value       = "${ece_elasticsearch_cluster.test_cluster.kibana_cluster_id}"
  description = "The ID of the Kibana cluster"
}

output "test_kibana_cluster_topology_0_memory_per_node" {
  value       = "${ece_elasticsearch_cluster.test_cluster.kibana.0.plan.0.cluster_topology.0.memory_per_node}"
  description = "The memory per node for the first topology element in the Kibana cluster"
}
```

## Development

### Requirements

* [Golang](https://golang.org/dl/) >= 1.11
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

go build -o ~/.terraform.d/plugins/darwin_amd64/terraform-provider-ece
```

## Contributing

1. Fork it ( https://github.com/Ascendon/terraform-provider-ece/fork )
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create a new Pull Request
