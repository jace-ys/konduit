package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/jace-ys/konduit/internal/exec"
	"github.com/jace-ys/konduit/internal/kustomize"
)

type KustomizeCmd struct {
	Manifests          *os.File `default:"-" arg:"" help:"Manifests file to run Kustomize on (use - for stdin)."`
	Dir                string   `required:"" help:"Directory to run Kustomize on."`
	PostRenderer       string   `help:"Original Helm post-renderer command to invoke."`
	PostRendererArgs   []string `help:"Original Helm post-renderer arguments to pass through."`
	KustomizeCommand   string   `default:"kustomize" help:"Kustomize command or path to an executable."`
	KustomizeBuildArgs []string `help:"Additional arguments to pass to Kustomize build."`
}

func (c *KustomizeCmd) Run(ctx context.Context, g *Globals) error {
	defer c.Manifests.Close()

	if c.Manifests != os.Stdin || c.Manifests.Fd() != os.Stdin.Fd() {
		return errors.New("manifests must be provided via stdin")
	}

	stat, err := c.Manifests.Stat()
	if err != nil {
		return fmt.Errorf("stat stdin: %w", err)
	}

	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return errors.New("no manifests from stdin")
	}

	manifests, err := kustomize.WriteManifests(c.Dir, c.Manifests)
	if err != nil {
		return fmt.Errorf("write manifests file: %w", err)
	}
	defer os.Remove(manifests)

	runner := exec.NewOSRunner()
	buildArgs := append([]string{"build", c.Dir}, c.KustomizeBuildArgs...)

	if c.PostRenderer == "" {
		if err := runner.Run(ctx, c.KustomizeCommand, buildArgs); err != nil {
			return fmt.Errorf("run kustomize: %w", err)
		}
		return nil
	}

	pr, pw := io.Pipe()
	errCh := make(chan error, 1)

	go func() {
		defer pr.Close()
		errCh <- runner.Run(ctx, c.PostRenderer, c.PostRendererArgs, exec.WithStdin(pr))
	}()

	if err := runner.Run(ctx, c.KustomizeCommand, buildArgs, exec.WithStdout(pw)); err != nil {
		pw.CloseWithError(err)
		return fmt.Errorf("run kustomize: %w", err)
	}
	pw.Close()

	if err := <-errCh; err != nil {
		return fmt.Errorf("run original post-renderer: %w", err)
	}

	return nil
}
