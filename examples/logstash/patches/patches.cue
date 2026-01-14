package patches

import (
	"github.com/jace-ys/konduit/examples/lib/k8s"
	"github.com/jace-ys/konduit/examples/logstash/lib:logstash"
)

_cluster: k8s.#Cluster & #Konduit.cluster

commonLabels: k8s.#Labels & {#cluster: _cluster}

secretGenerator: [{
	name: logstash.#ElasticsearchMonitoringRef.secretName
	literals: [
		"url=https://monitoring.elasticsearch.konduit.io",
		"username=\(logstash.#ElasticsearchMonitoringRef.user)",
		"password=\(logstash.#ElasticsearchMonitoringRef.password)",
	]
}]

generatorOptions: disableNameSuffixHash: true
