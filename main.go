package main

/*
Modeled after structure & functionality as found here: https://github.com/phillbaker/terraform-provider-elasticsearch
*/

import (
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: Provider,
	})
}
