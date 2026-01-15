package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jace-ys/konduit/pkg/cueval"
	"github.com/jace-ys/konduit/pkg/konduit"
)

type CUECmd struct {
	Show bool `help:"Print the resulting Helm invocation, with evaluated values and patches."`

	Values  []string `short:"v" help:"Helm values files to be evaluated by CUE."`
	Patches []string `short:"p" help:"Kustomize patches files to be evaluated by CUE."`
	Scopes  []string `short:"s" sep:"none" help:"JSON/YAML data (or @filename) to inject under the #Konduit definition."`

	Args        []string `arg:"" passthrough:"partial" help:"Arguments after the leading -- are passed through to Helm."`
	HelmCommand string   `help:"Helm command or path to an executable."`

	CUEBaseDir    string `help:"Base directory for import path resolution. If empty, the current directory is used."`
	CUEModuleRoot string `help:"Directory that contains the cue.mod directory and packages."`

	Strict bool `help:"Disallow using evaluated and static configuration at the same time."`
}

func (c *CUECmd) Run(ctx context.Context, g *Globals) error {
	if c.Args[0] != "--" {
		return errors.New("must use -- to pass through Helm arguments")
	}

	eval := konduit.NewCUEEvaluator(
		cueval.WithScopes(c.Scopes...),
		cueval.WithLoadDir(c.CUEBaseDir),
		cueval.WithLoadModuleRoot(c.CUEModuleRoot),
	)

	opts := []konduit.Option{
		konduit.WithEvaluator(eval),
		konduit.WithModeStrict(c.Strict),
	}

	if len(c.Patches) > 0 {
		opts = append(opts, konduit.WithPatches(c.Patches))
	}

	if c.HelmCommand != "" {
		opts = append(opts, konduit.WithHelmCommand(c.HelmCommand))
	}

	k, err := konduit.New(c.Args[1:], c.Values, opts...)
	if err != nil {
		return fmt.Errorf("init: %w", err)
	}

	if c.Show {
		cmd, err := k.Construct()
		if err != nil {
			return fmt.Errorf("construct invocation: %w", err)
		}

		if err := json.NewEncoder(g.Stdout).Encode(cmd); err != nil {
			return fmt.Errorf("encode invocation: %w", err)
		}

		return nil
	}

	if err := k.Execute(ctx); err != nil {
		return fmt.Errorf("execute: %w", err)
	}

	return nil
}
