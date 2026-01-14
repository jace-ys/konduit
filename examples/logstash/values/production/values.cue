package values

import "github.com/jace-ys/konduit/examples/logstash/lib:logstash"

logstash & {
	podTemplate: spec: containers: [{
		resources: requests: {
			cpu:    "100m"
			memory: "256Mi"
		}
	}]

	#Topics: [{
		name: "logs.topic-01"
		pipeline: {
			"config.debug":     true
			"pipeline.workers": 8
		}
	}, {
		name: "logs.topic-02"
		pipeline: {
			"config.debug":        true
			"pipeline.batch.size": 500
		}
	}]
}
