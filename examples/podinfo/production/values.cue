package values

import "github.com/jace-ys/konduit/examples/lib/k8s"

resources: k8s.#ResourceRequirements & {
	requests: cpu:    "100m"
	requests: memory: "256Mi"
}
