package values

import "github.com/jace-ys/konduit/examples/logstash/lib:logstash"

logstash & {
	podTemplate: spec: containers: [{
		resources: requests: {
			cpu:    "25m"
			memory: "64Mi"
		}
	}]

	#Topics: [{
		name: "logs.topic-01"
		pipeline: {
			"config.debug":     true
			"pipeline.workers": 2
		}
	}]
}
