# Usage

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Command Reference](#command-reference)
- [Values](#values)
- [Patches](#patches)
- [Scopes](#scopes)
- [CUE Modules](#cue-modules)
- [Post-Renderer Chaining](#post-renderer-chaining)
- [Debugging](#debugging)
- [Go SDK](#go-sdk)
- [Examples](#examples)

---

## Prerequisites

When running Konduit locally, the following tools must be installed and available in your `PATH`:

| Tool | Required | Tested Version |
|------|----------|----------------|
| [Helm](https://helm.sh/docs/intro/install/) | Yes | 3.19+ (Helm 4 not yet supported) |
| [Kustomize](https://kubectl.docs.kubernetes.io/installation/kustomize/) | If using `-p` patches | 5.8+ |
| [CUE](https://cuelang.org/docs/introduction/installation/) | No (for development purposes) | 0.15+ |

## Installation

```shell
# Homebrew
brew install jace-ys/tap/konduit

# Helm plugin
helm plugin install https://github.com/jace-ys/konduit

# Docker
docker pull ghcr.io/jace-ys/konduit:latest
```

See the [releases page](https://github.com/jace-ys/konduit/releases) for precompiled binaries.

### Helm Plugin Usage

When installed as a Helm plugin, use `helm konduit` instead of `konduit`:

```shell
# Standalone
konduit cue -v values.cue -- template my-release ./chart

# As Helm plugin
helm konduit cue -v values.cue -- template my-release ./chart
```

All flags and features work identically. Installing the plugin automatically downloads the correct binary for your platform.

---

## Quick Start

```cue
// values.cue
package values

replicas: 3
image: {
    repository: "nginx"
    tag:        "1.25"
}
```

```shell
konduit cue -v values.cue -- template my-release ./chart
```

Konduit evaluates the CUE file to YAML and passes it to Helm.

---

## Command Reference

### `konduit cue`

```shell
Usage: konduit cue <args> ... [flags]

Run Helm with CUE evaluation of Helm values and Kustomize patches.

Arguments:
  <args> ...    Arguments after the leading -- are passed through to Helm.

Flags:
  -h, --help                      Show context-sensitive help.
      --log.level="info"          Configure the log level ($LOG_LEVEL).
      --log.format="text"         Configure the log format ($LOG_FORMAT).

      --show                      Print the resulting Helm invocation, with evaluated values and patches.
  -v, --values=VALUES,...         Helm values files to be evaluated by CUE.
  -p, --patches=PATCHES,...       Kustomize patches files to be evaluated by CUE.
  -s, --scopes=SCOPES             JSON/YAML data (or @filename) to inject under the #Konduit definition.
      --helm-command=STRING       Helm command or path to an executable.
      --cue-base-dir=STRING       Base directory for import path resolution. If empty, the current directory is used.
      --cue-module-root=STRING    Directory that contains the cue.mod directory and packages.
      --strict                    Disallow using evaluated and static configuration at the same time.
```

---

## Values

Values files define [Helm chart values](https://helm.sh/docs/chart_template_guide/values_files/). All values (CUE and YAML) should be passed to Konduit via `-v`. Konduit handles them and passes them to Helm accordingly.

### CUE Files

CUE files are [unified](https://cuelang.org/docs/tour/basics/unification/) into a single evaluated YAML file:

```shell
konduit cue -v base.cue -v production.cue -- template my-release ./chart
```

### YAML Files

YAML files take precedence over CUE files and override each other in order, following standard Helm behavior:

```shell
konduit cue -v values.cue -v overrides.yaml -- template my-release ./chart
```

In this example, `overrides.yaml` values override matching keys from the evaluated CUE output.

> **Note:** While Helm's `-f`/`--values` flags after `--` are still handled correctly, passing all values to Konduit via `-v` is recommended for clarity.

---

## Patches

Patches extend manifests from Helm charts using Kustomize - use them when a chart doesn't expose the values or templates you need.

In Konduit's context, "patches" refer to [kustomization](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/) files containing [built-in generators or transformers](https://kubectl.docs.kubernetes.io/references/kustomize/builtins/).

### Supported Format

Patches must use **kustomization field syntax** — the same fields you'd put in a `kustomization.yaml` file. See the [kustomization reference](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/) for the full list of supported fields.

> **Note:** Standalone transformer configurations are **not currently supported**. Use the equivalent kustomization fields instead — for example, use `commonLabels` instead of `LabelTransformer`.

### How It Works

This is done via a Helm [post-renderer](https://helm.sh/docs/v3/topics/advanced/#post-rendering), similar to the [example](https://github.com/thomastaylor312/advanced-helm-demos/blob/master/post-render/kustomize/kustomize) provided in their docs:

1. Konduit writes patches to a temp `kustomization.yaml` file
1. Konduit invokes Helm with the [`konduit kustomize` post-renderer](../cmd/konduit/kustomize.go)
1. Helm command renders manifests from charts to stdout
1. Konduit post-renderer writes manifests to the temp directory
1. Konduit invokes Kustomize to apply the `kustomization.yaml`

### Example

```cue
// patches.cue
package patches

commonLabels: {
    "app.kubernetes.io/managed-by": "konduit"
}

patches: [{
    target: kind: "Deployment"
    patch: """
        - op: add
          path: /spec/template/spec/securityContext
          value:
            runAsNonRoot: true
        """
}]
```

```shell
konduit cue -v values.cue -p patches.cue -- template my-release ./chart
```

---

## Scopes

Scopes allow you to inject external data into CUE under the `#Konduit` [definition](https://cuelang.org/docs/tour/basics/definitions/).

### Passing Scopes

```shell
# From a file
konduit cue -s @cluster.json -v values.cue -- template my-release ./chart

# Inline JSON/YAML
konduit cue -s '{"environment": "production"}' -v values.cue -- template my-release ./chart

# Multiple scopes (unioned by CUE)
konduit cue -s @cluster.json -s @secrets.yaml -v values.cue -- template my-release ./chart
```

### Using Scopes in CUE

```cue
// values.cue
package values

image: repository: "\(#Konduit.cluster.registry)/my-app"

podLabels: {
    "kubernetes.konduit.io/cluster":     #Konduit.cluster.name
    "kubernetes.konduit.io/environment": #Konduit.cluster.tags.environment
    "kubernetes.konduit.io/region":      #Konduit.cluster.tags.region
}
```

### Example Scope File

```json
{
  "cluster": {
    "name": "cluster-01",
    "registry": "ghcr.io",
    "tags": {
      "environment": "production",
      "region": "us-east-1"
    }
  }
}
```

### With Patches

Scopes can also be used in patches. A useful pattern is injecting secrets into Kustomize's `secretGenerator`:

```cue
// patches.cue
package patches

secretGenerator: [{
    name: "app-secrets"
    literals: [
        "API_KEY=\(#Konduit.secrets.API_KEY)",
    ]
}]
```

```shell
konduit cue -v values.cue -p patches.cue -s '{"secrets": {"API_KEY": "abc123"}}' -- template my-release ./chart
```

This keeps sensitive values out of configuration files.

---

## CUE Modules

For projects that use imports, set up a [CUE module](https://cuetorials.com/first-steps/modules-and-packages/):

```shell
cue mod init github.com/owner/repo
```

Then specify the base directory (containing `cue.mod/`) for resolving imports:

```shell
# Directory structure:
# ├── lib/
# │   ├── cue.mod/
# │   └── k8s/
# │       └── labels.cue
# └── app/
#     └── values.cue

konduit cue \
    --cue-base-dir ./lib \
    -v ./app/values.cue \
    -- template my-release ./chart
```

### Importing Libraries

Use [`import`](https://cuelang.org/docs/tour/packages/imports/) to reference packages from your module:

```cue
// app/values.cue
package values

import "github.com/owner/repo/k8s"

labels: k8s.#Labels & {#cluster: #Konduit.cluster}
```

---

## Post-Renderer Chaining

Konduit chains with other Helm post-renderers:

```shell
konduit cue \
    -v values.cue \
    -p patches.cue \
    -- template my-release ./chart \
    --post-renderer ./my-renderer \
    --post-renderer-args arg1
```

---

## Debugging

### Dry Run

Use `--show` to perform a dry run and inspect what Konduit will pass to Helm:

```shell
konduit cue --show -v values.cue -- template my-release ./chart | yq -P
```

Output includes:
- `command`: Helm executable
- `args`: Arguments to pass to Helm
- `evaluatedValues`: CUE evaluation result
- `evaluatedPatches`: Patch evaluation result

### Validate CUE

```shell
cue vet values.cue
cue export values.cue --out yaml
```

---

## Go SDK

Konduit can also be used programmatically via its Go library. See the API documentation for [`pkg/konduit`](https://pkg.go.dev/github.com/jace-ys/konduit/pkg/konduit) and [`pkg/cueval`](https://pkg.go.dev/github.com/jace-ys/konduit/pkg/cueval).

### Installation

```shell
go get github.com/jace-ys/konduit
```

### Basic Usage

```go
package main

import (
    "context"

    "github.com/jace-ys/konduit/pkg/cueval"
    "github.com/jace-ys/konduit/pkg/konduit"
)

func main() {
    // Create a CUE evaluator with options
    eval := konduit.NewCUEEvaluator(
        cueval.WithScopes("@cluster.json"),
        cueval.WithLoadDir("./lib"),
    )

    // Create a Konduit instance
    k, err := konduit.New(
        []string{"template", "my-release", "./chart"},
        []string{"values.cue", "production.cue"},
        konduit.WithEvaluator(eval),
        konduit.WithPatches([]string{"patches.cue"}),
    )
    if err != nil {
        panic(err)
    }

    // Execute runs Helm with evaluated values and patches
    if err := k.Execute(context.Background()); err != nil {
        panic(err)
    }
}
```

### Dry Run

Use `Construct()` to inspect the invocation without executing:

```go
inv, err := k.Construct()
if err != nil {
    panic(err)
}

fmt.Println("Command:", inv.Command)
fmt.Println("Args:", inv.Args)
fmt.Println("Evaluated Values:", inv.EvaluatedValues.ResultYAML)
fmt.Println("Evaluated Patches:", inv.EvaluatedPatches.ResultYAML)
```


### Custom Evaluator

Implement the `Evaluator` interface for custom evaluators:

```go
type Evaluator interface {
    Evaluate(files []string) ([]byte, error)
    SupportedFileExt() string
}
```

---

## Examples

### Basic

```shell
# Single values file
konduit cue -v values.cue -- template my-release ./chart

# Install with values
konduit cue -v values.cue -- install my-release ./chart

# Upgrade with multiple values
konduit cue -v base.cue -v production.cue -- upgrade my-release ./chart
```

### With Patches

```shell
# With values and patches
konduit cue -v values.cue -p patches.cue -- template my-release ./chart
```

### With Scopes

```shell
# Environment-specific configuration
konduit cue \
    -v ./app/values.cue \
    -p ./app/patches.cue \
    -s @./clusters/production.json \
    -- install my-release ./chart
```

### Full Example

```shell
# Directory structure:
# ├── lib/
# │   ├── cue.mod/
# │   └── k8s/
# │       └── labels.cue
# ├── app/
# │   ├── values.cue
# │   ├── patches.cue
# │   └── production/
# │       └── values.cue
# ├── clusters/
# │   └── production.json
# └── charts/
#     └── my-app/

konduit cue \
    --cue-base-dir ./lib \
    -v ./app/values.cue \
    -v ./app/production/values.cue \
    -p ./app/patches.cue \
    -s @./clusters/production.json \
    -- install my-app ./charts/my-app \
    --namespace production
```

See the [examples directory](../examples/README.md) for complete working examples.