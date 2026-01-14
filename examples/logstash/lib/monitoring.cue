package logstash

#ElasticsearchMonitoringRef: {
	user:       "monitoring"
	password:   #Konduit.secrets.ELASTICSEARCH_MONITORING_ES_PASSWORD
	secretName: "elasticsearch-monitoring"
}

let elasticsearchMonitoringRef = {
	secretName: #ElasticsearchMonitoringRef.secretName
}

monitoring: {
	metrics: elasticsearchRefs: [elasticsearchMonitoringRef]
	logs: elasticsearchRefs: [elasticsearchMonitoringRef]
}
