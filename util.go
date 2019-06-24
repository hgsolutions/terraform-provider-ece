package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

/*
NOTE: This would need some refactoring to work for our needs, if it's even needed.
*/

func diffSuppressClusterSettings(k, old, new string, d *schema.ResourceData) bool {
	var oo, no interface{}
	if err := json.Unmarshal([]byte(old), &oo); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &no); err != nil {
		return false
	}

	if om, ok := oo.(map[string]interface{}); ok {
		normalizeClusterSettings(om)
	}

	if nm, ok := no.(map[string]interface{}); ok {
		normalizeClusterSettings(nm)
	}

	return reflect.DeepEqual(oo, no)
}

func normalizeClusterSettings(tpl map[string]interface{}) {
	//delete(tpl, "version") // Shouldn't exist in the JSON.
	if settings, ok := tpl["settings"]; ok {
		if settingsMap, ok := settings.(map[string]interface{}); ok {
			tpl["settings"] = normalizedClusterSettings(settingsMap)
		}
	}
}

func normalizedClusterSettings(settings map[string]interface{}) map[string]interface{} {
	f := flattenMap(settings)
	for k, v := range f {
		f[k] = fmt.Sprintf("%v", v)
		if !strings.HasPrefix(k, "index.") {
			f["index."+k] = fmt.Sprintf("%v", v)
			delete(f, k)
		}
	}

	return f
}

func flattenMap(m map[string]interface{}) map[string]interface{} {
	f := make(map[string]interface{})
	for k, v := range m {
		if vm, ok := v.(map[string]interface{}); ok {
			fm := flattenMap(vm)
			for k2, v2 := range fm {
				f[k+"."+k2] = v2
			}
		} else {
			f[k] = v
		}
	}

	return f
}
