package patches

import "github.com/jace-ys/konduit/examples/lib/k8s"

_cluster: k8s.#Cluster & #Konduit.cluster

commonLabels: k8s.#Labels & {#cluster: _cluster}
