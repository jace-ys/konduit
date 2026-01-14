package values

import "github.com/jace-ys/konduit/examples/lib/k8s"

_cluster: k8s.#Cluster & #Konduit.cluster

podAnnotations: k8s.#Annotations & {#cluster: _cluster}
