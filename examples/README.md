# Examples

This directory contains example configurations demonstrating how to use Konduit with off-the-shelf Helm charts.

## Directory Structure

```
examples/
├── cue.mod/              # CUE module
├── data/                 # Environment-specific scope data
│   ├── development.json
│   └── production.json
├── lib/                  # Shared CUE libraries
│   └── k8s/              # Kubernetes-related definitions
├── podinfo/              # Simple example
└── logstash/             # Advanced example
```

## Shared Libraries

Reusable CUE definitions imported by both examples:

| File | Purpose |
|------|---------|
| `lib/k8s/cluster.cue` | Cluster schema for scope validation |
| `lib/k8s/metadata.cue` | Standardized labels and annotations |
| `lib/k8s/container.cue` | Best-practices for container memory and CPU |

---

## Podinfo

A simple example demonstrating basic Konduit usage.

### Features

1. Uses CUE to evaluate the base `values.cue` with environment-specific `values.cue`
1. Injects environment-specific cluster data into the CUE evaluation via scopes
1. Imports reusable CUE libraries for enforcing common patterns and constraints
1. Applies Kustomize patches from `patches.cue` that extends the Helm chart
1. Chains two additional Helm post-renderers (`kustomize-1`, `kustomize-2`)

### Run

```shell
# Development
konduit cue \
    -v podinfo/values.cue \
    -v podinfo/development/values.cue \
    -p podinfo/patches.cue \
    -s @data/development.json \
    -- \
    template podinfo podinfo/podinfo-6.9.4.tgz \
    --post-renderer ../hack/post-render-chain \
    --post-renderer-args kustomize-1 \
    --post-renderer-args kustomize-2
```

```shell
# Production
konduit cue \
    -v podinfo/values.cue \
    -v podinfo/production/values.cue \
    -p podinfo/patches.cue \
    -s @data/production.json \
    -- \
    template podinfo podinfo/podinfo-6.9.4.tgz \
    --post-renderer ../hack/post-render-chain \
    --post-renderer-args kustomize-1 \
    --post-renderer-args kustomize-2
```

---

## Logstash

A full-featured example demonstrating advanced Konduit + CUE patterns.

See the [integration tests](../tests/testdata/) for expected output.

### Features

1. Combines both CUE + YAML for Helm values and Kustomize patches
1. Defines a CUE library for Logstash values that can be reused across environments
1. Uses Kustomize to generate Kubernetes secrets from secret data injected via scopes
1. Defines relationships between values and can also be referenced by patches (secret name)
1. Generates configuration dynamically (Logstash pipelines) using CUE expressions

### Run

```shell
# Development
konduit cue \
    -v logstash/values/development/values.cue \
    -v logstash/values/development/values.yaml \
    -p logstash/patches/patches.cue \
    -p logstash/patches/patches.yaml \
    -s @data/development.json \
    -s '{"secrets": {"ELASTICSEARCH_MONITORING_ES_PASSWORD": "secret123"}}' \
    -- \
    template logstash logstash/eck-logstash-0.17.0.tgz
```

```shell
# Production
konduit cue \
    -v logstash/values/production/values.cue \
    -v logstash/values/production/values.yaml \
    -p logstash/patches/patches.cue \
    -p logstash/patches/patches.yaml \
    -s @data/production.json \
    -s '{"secrets": {"ELASTICSEARCH_MONITORING_ES_PASSWORD": "secret123"}}' \
    -- \
    template logstash logstash/eck-logstash-0.17.0.tgz
```
