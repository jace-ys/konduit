package konduit

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/jace-ys/konduit/internal/exec"
)

const DefaultHelmCommand = "helm"

type Instance struct {
	HelmCommand string
	HelmArgs    []string

	PostRenderer     string
	PostRendererArgs []string

	Values           []string
	ValuesToEvaluate []string

	Patches           []string
	PatchesToEvaluate []string
	patchesOpt        []string

	dir    string
	strict bool

	evaluator Evaluator
	runner    exec.Runner
}

//nolint:cyclop
func New(args []string, values []string, opts ...Option) (*Instance, error) {
	if len(args) == 0 {
		return nil, errors.New("no arguments provided to Helm")
	}

	instance := &Instance{
		HelmCommand: DefaultHelmCommand,
		evaluator:   NewNoopEvaluator(),
		runner:      exec.NewOSRunner(),
	}

	for _, opt := range opts {
		opt.Apply(instance)
	}

	for _, value := range values {
		if filepath.Ext(value) == instance.evaluator.SupportedFileExt() {
			instance.ValuesToEvaluate = append(instance.ValuesToEvaluate, value)
		} else {
			instance.Values = append(instance.Values, value)
		}
	}

	for _, patch := range instance.patchesOpt {
		if filepath.Ext(patch) == instance.evaluator.SupportedFileExt() {
			instance.PatchesToEvaluate = append(instance.PatchesToEvaluate, patch)
		} else {
			instance.Patches = append(instance.Patches, patch)
		}
	}

	parseHelmArgs(instance, args)

	if instance.strict {
		if len(instance.ValuesToEvaluate) > 0 && len(instance.Values) > 0 {
			return nil, errors.New("strict mode enabled; can't use evaluated and static values at the same time")
		}
		if len(instance.PatchesToEvaluate) > 0 && len(instance.Patches) > 0 {
			return nil, errors.New("strict mode enabled; can't use evaluated and static patches at the same time")
		}
	}

	return instance, nil
}

type argKind int

const (
	argKindNone argKind = iota
	argKindValues
	argKindPostRenderer
	argKindPostRendererArgs
)

//nolint:cyclop
func parseHelmArgs(i *Instance, args []string) {
	for n := 0; n < len(args); n++ {
		arg := args[n]

		var val string
		var kind argKind

		switch {
		// Values: -f, --values
		case strings.HasPrefix(arg, "--values="):
			val, kind = strings.TrimPrefix(arg, "--values="), argKindValues
		case strings.HasPrefix(arg, "-f="):
			val, kind = strings.TrimPrefix(arg, "-f="), argKindValues
		case arg == "-f", arg == "--values":
			if n+1 < len(args) {
				n++
				val, kind = args[n], argKindValues
			}

		// Post-renderer: --post-renderer
		case strings.HasPrefix(arg, "--post-renderer="):
			val, kind = strings.TrimPrefix(arg, "--post-renderer="), argKindPostRenderer
		case arg == "--post-renderer":
			if n+1 < len(args) {
				n++
				val, kind = args[n], argKindPostRenderer
			}

		// Post-renderer args: --post-renderer-arg
		case strings.HasPrefix(arg, "--post-renderer-args="):
			val, kind = strings.TrimPrefix(arg, "--post-renderer-args="), argKindPostRendererArgs
		case arg == "--post-renderer-args":
			if n+1 < len(args) {
				n++
				val, kind = args[n], argKindPostRendererArgs
			}

		default:
			val, kind = arg, argKindNone
		}

		switch kind {
		case argKindValues:
			i.Values = append(i.Values, val)
		case argKindPostRenderer:
			i.PostRenderer = val
		case argKindPostRendererArgs:
			i.PostRendererArgs = append(i.PostRendererArgs, val)
		case argKindNone:
			i.HelmArgs = append(i.HelmArgs, arg)
		}
	}
}
