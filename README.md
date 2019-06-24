# terraform-provider-ece

** IN PROGRESS **

Terraform provider for provisioning Elastic Cloud Enterprise (ECE) Elasticsearch clusters, compatible with v2.2 of ECE. 

Based on work by Phillip Baker on the [terraform-provider-elasticsearch](https://github.com/phillbaker/terraform-provider-elasticsearch).

## Installation

TODO: Download a binary, and put it in a good spot on your system. Then update your `~/.terraformrc` to refer to the binary:

```hcl
providers {
  ece = "/path/to/terraform-provider-ece"
}
```

See [the docs for more information](https://www.terraform.io/docs/plugins/basics.html).

## Usage

```tf
provider "ece" {
    url = "http://ec2-18-234-124-116.compute-1.amazonaws.com:12400"
    user = ""
    password = ""
    insecure = true # to bypass certificate check
}

resource "ece_cluster" "test_cluster" {
  name = "My First Cluster"
  elasticsearch_version = "7.1.0",
  memory_per_node" : 2048,
  node_count_per_zone" : 1,
  node_type" : {
        "data" : true,
        "ingest" : true,
        "master" : true,
        "ml" : true
  },
  "zone_count" : 1
}
```

## Development

### Requirements

* [Golang](https://golang.org/dl/) >= 1.7
* [Glide](https://github.com/Masterminds/glide)

```
# Ensure that this folder is at the following location: `${GOPATH}/src/github.com/Ascendon/terraform-provider-ece`

cd $GOPATH/src/github.com/Ascendon/terraform-provider-ece

glide install
go build -o releases/terraform-provider-ece
```

## Contributing

1. Fork it ( https://github.com/Ascendon/terraform-provider-ece/fork )
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create a new Pull Request
