package k8s

import "encoding/json"

#Labels: [string]: string
#Labels: {
	#cluster: #Cluster

	"kubernetes.konduit.io/cluster":     #cluster.name
	"kubernetes.konduit.io/environment": #cluster.tags.environment
	"kubernetes.konduit.io/region":      #cluster.tags.region
}

#Annotations: [string]: string
#Annotations: {
	#cluster: #Cluster

	if #cluster.attributes.vault != _|_ {
		if #cluster.attributes.vault.enabled {
			"vault.hashicorp.com/agent-inject": "true"
		}
	}

	#Filebeat: enabled: *true | bool
	if #Filebeat.enabled {
		"co.elastic.logs/enabled": "true"
	}

	#Datadog: {
		enabled:   *false | bool
		container: string
		check: {...}
	}
	if #Datadog.enabled {
		"ad.datadoghq.com/\(#Datadog.container).checks": json.Indent(json.Marshal(#Datadog.check), "", "  ")
	}
}
