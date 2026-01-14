package kustomize

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
	kustomize "sigs.k8s.io/kustomize/api/types"
)

const (
	ManifestsFile     = "manifests.yaml"
	KustomizationFile = "kustomization.yaml"
)

func MakeDefinition(patches ...[]byte) (*kustomize.Kustomization, error) {
	k := new(kustomize.Kustomization)
	opts := []yaml.DecodeOption{yaml.DisallowUnknownField()}

	for _, patch := range patches {
		if err := yaml.UnmarshalWithOptions(patch, k, opts...); err != nil {
			return nil, fmt.Errorf("decode patch: %w", err)
		}
	}

	k.Resources = append(k.Resources, ManifestsFile)
	k.FixKustomization()

	return k, nil
}

func WriteKustomization(dir string, k *kustomize.Kustomization) (string, error) {
	filename := filepath.Join(dir, KustomizationFile)

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	if err := yaml.NewEncoder(file).Encode(k); err != nil {
		return "", fmt.Errorf("encode to file: %w", err)
	}

	return filename, nil
}

func WriteManifests(dir string, manifests io.Reader) (string, error) {
	filename := filepath.Join(dir, ManifestsFile)

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}

	if _, err := io.Copy(file, manifests); err != nil {
		return "", fmt.Errorf("copy to file: %w", err)
	}
	defer file.Close()

	return filename, nil
}
