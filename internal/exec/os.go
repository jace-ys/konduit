package exec

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type OSRunner struct{}

func NewOSRunner() *OSRunner {
	return &OSRunner{}
}

func (r *OSRunner) Run(ctx context.Context, command string, args []string, opts ...Option) error {
	executable, err := exec.LookPath(command)
	if err != nil {
		return fmt.Errorf("find executable %q: %w", command, err)
	}

	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	options := new(Options)
	for _, opt := range opts {
		opt.Apply(options)
	}

	if options.stdin != nil {
		cmd.Stdin = options.stdin
	}

	if options.stdout != nil {
		cmd.Stdout = options.stdout
	}

	if options.dir != "" {
		cmd.Dir = options.dir
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("exec command: %w", err)
	}

	return nil
}
