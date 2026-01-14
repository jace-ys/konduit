package logstash

#Topic: {
	name!: string
	pipeline: {...}
}

#Topics: [...#Topic]

#Pipeline: {
	"pipeline.id"!:   string
	"config.string"!: string
	...
}

pipelines: [...#Pipeline]
pipelines: [
	for topic in #Topics {
		topic.pipeline

		"pipeline.id":   topic.name
		"config.string": #"""
			input {
			  kafka {
			    bootstrap_servers => "${KAFKA_BOOTSTRAP_SERVERS}"
			    topics => ["\#(topic.name)"]
			    group_id => "logstash/\#(topic.name)"
			    codec => "json"
			  }
			}
			output {
			  elasticsearch {
			    hosts => [ "${ELASTICSEARCH_ES_HOSTS}" ]
			    user => "${ELASTICSEARCH_ES_USER}"
			    password => "${ELASTICSEARCH_ES_PASSWORD}"
			    ssl_certificate_authorities => "${ELASTICSEARCH_ES_SSL_CERTIFICATE_AUTHORITY}"
			  }
			}
			"""#
	},
]
