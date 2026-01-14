package konduit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jace-ys/konduit/internal/kustomize"
)

const ValuesFile = "evaluated.yaml"

type Invocation struct {
	Command          string      `json:"command"`
	Args             []string    `json:"args"`
	EvaluatedValues  *Evaluation `json:"evaluatedValues"`
	Values           []string    `json:"values,omitempty"`
	EvaluatedPatches *Evaluation `json:"evaluatedPatches"`
	Patches          []string    `json:"patches,omitempty"`
}

func (i *Instance) Construct() (*Invocation, error) {
	cmd := &Invocation{
		Command:          i.HelmCommand,
		Args:             i.constructHelmArgs(),
		Values:           i.Values,
		Patches:          i.Patches,
		EvaluatedValues:  &Evaluation{Files: i.ValuesToEvaluate},
		EvaluatedPatches: &Evaluation{Files: i.PatchesToEvaluate},
	}

	if len(i.ValuesToEvaluate) > 0 {
		result, err := i.evaluator.Evaluate(i.ValuesToEvaluate)
		if err != nil {
			return nil, fmt.Errorf("evaluate values: %w", err)
		}
		cmd.EvaluatedValues.ResultYAML = string(result)
	}

	if len(i.PatchesToEvaluate) > 0 {
		result, err := i.evaluator.Evaluate(i.PatchesToEvaluate)
		if err != nil {
			return nil, fmt.Errorf("evaluate patches: %w", err)
		}
		cmd.EvaluatedPatches.ResultYAML = string(result)
	}

	return cmd, nil
}

func (i *Instance) constructHelmArgs() []string {
	args := i.HelmArgs

	if len(i.ValuesToEvaluate) > 0 {
		args = append(args, "--values", filepath.Join(i.dir, ValuesFile))
	}

	for _, value := range i.Values {
		args = append(args, "--values", value)
	}

	if len(i.Patches) > 0 || len(i.PatchesToEvaluate) > 0 {
		konduitCommand, err := os.Executable()
		if err != nil {
			konduitCommand = "konduit"
		}

		args = append(args,
			"--post-renderer", konduitCommand,
			"--post-renderer-args", "kustomize",
			"--post-renderer-args", "--dir",
			"--post-renderer-args", i.dir,
		)

		if i.PostRenderer != "" {
			args = append(args, "--post-renderer-args", "--post-renderer")
			args = append(args, "--post-renderer-args", i.PostRenderer)

			for _, arg := range i.PostRendererArgs {
				args = append(args, "--post-renderer-args", "--post-renderer-args")
				args = append(args, "--post-renderer-args", arg)
			}
		}
	} else if i.PostRenderer != "" {
		args = append(args, "--post-renderer", i.PostRenderer)

		for _, arg := range i.PostRendererArgs {
			args = append(args, "--post-renderer-args", arg)
		}
	}

	return args
}

func (i *Instance) Execute(ctx context.Context) error {
	if i.dir == "" {
		dir, err := os.MkdirTemp("", "konduit-*")
		if err != nil {
			return fmt.Errorf("create tmp dir: %w", err)
		}
		defer os.RemoveAll(dir)
		i.dir = dir
	}

	inv, err := i.Construct()
	if err != nil {
		return fmt.Errorf("construct invocation: %w", err)
	}

	if err := inv.prepareHelm(i.dir); err != nil {
		return err
	}

	if err := inv.prepareKustomize(i.dir); err != nil {
		return err
	}

	if err := i.runner.Run(ctx, inv.Command, inv.Args); err != nil {
		return fmt.Errorf("run invocation: %w", err)
	}

	return nil
}

func (i *Invocation) prepareHelm(dir string) error {
	if len(i.EvaluatedValues.ResultYAML) > 0 {
		filename := filepath.Join(dir, ValuesFile)

		file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			return fmt.Errorf("create evaluated values file: %w", err)
		}
		defer file.Close()

		if _, err := file.WriteString(i.EvaluatedValues.ResultYAML); err != nil {
			return fmt.Errorf("write evaluated values file: %w", err)
		}
	}

	return nil
}

func (i *Invocation) prepareKustomize(dir string) error {
	patches := make([][]byte, 0)

	if len(i.EvaluatedPatches.ResultYAML) > 0 {
		patches = append(patches, []byte(i.EvaluatedPatches.ResultYAML))
	}

	for _, patch := range i.Patches {
		content, err := os.ReadFile(patch)
		if err != nil {
			return fmt.Errorf("read patch file: %w", err)
		}
		patches = append(patches, content)
	}

	kustomization, err := kustomize.MakeDefinition(patches...)
	if err != nil {
		return fmt.Errorf("define kustomization: %w", err)
	}

	if _, err := kustomize.WriteKustomization(dir, kustomization); err != nil {
		return fmt.Errorf("write kustomization file: %w", err)
	}

	return nil
}
