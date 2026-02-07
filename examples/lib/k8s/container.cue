package k8s

import corev1 "cue.dev/x/k8s.io/api/core/v1"

#ResourceRequirements: X={
	corev1.#ResourceRequirements
	limits: memory: X.requests.memory
	limits: cpu?:   _|_
}
