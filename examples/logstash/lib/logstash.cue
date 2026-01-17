package logstash

import (
	corev1 "cue.dev/x/k8s.io/api/core/v1"

	"github.com/jace-ys/konduit/examples/lib/k8s"
)

let app = "logstash"

_cluster: k8s.#Cluster & #Konduit.cluster

version: string | *"9.2.3"

elasticsearchRefs: [{
	name:      "elasticsearch"
	namespace: name
}]

image: "\(_cluster.attributes.images.registry)/\(app):\(version)"

podTemplate: corev1.#PodTemplateSpec
podTemplate: {
	metadata: {
		labels: k8s.#Labels & {#cluster: _cluster}
		annotations: k8s.#Annotations & {#cluster: _cluster}
		annotations: #Datadog: {
			enabled:   true
			container: app
			check: logstash: instances: [{
				url: "http://%%host%%:9600"
			}]
		}
	}
	spec: containers: [{
		name:      app
		resources: k8s.#ResourceRequirements
	}]
}
