package k8s

#Cluster: {
	name:       string
	tags:       #ClusterTags
	attributes: #ClusterAttributes
}

#ClusterTags: {
	environment: string
	region:      string
}

#ClusterAttributes: {
	images: registry: string
	vault?: enabled:  *false | bool
}
