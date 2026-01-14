package k8s

import corev1 "k8s.io/api/core/v1"

#ResourceRequirements: corev1.#ResourceRequirements & {
	requests: {
		memory: string
		cpu:    string
	}
	limits: memory: requests.memory
}
