# terraform-provider-ece

Terraform provider for provisioning Elastic Cloud Enterprise (ECE) Elasticsearch clusters, compatible with v2.2 of ECE. 

Based on work by Phillip Baker: [terraform-provider-elasticsearch](https://github.com/phillbaker/terraform-provider-elasticsearch).

## Installation

Build or download a binary from the releases folder and put it in your Terraform user plugins directory.

See [the docs for more information](https://www.terraform.io/docs/plugins/basics.html).

## Usage

**NOTE:** Only a subset of the ECE API configuration parameters are currently implemented. See the `CreateElasticsearchClusterRequest` structure in the `ece_api_structures.go` file for the currently supported parameters.

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
