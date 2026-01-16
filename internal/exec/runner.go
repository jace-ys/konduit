package exec

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type OSRunner struct{}

func NewOSRunner() *OSRunner {
	return &OSRunner{}
}

func (r *OSRunner) Run(ctx context.Context, command string, args []string, opts ...RunOption) error {
	executable, err := exec.LookPath(command)
	if err != nil {
		return fmt.Errorf("find executable %q: %w", command, err)
	}

	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	options := new(runOptions)
	for _, opt := range opts {
		opt.apply(options)
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

type runOptions struct {
	stdin  io.Reader
	stdout io.Writer
	dir    string
}

type RunOption interface {
	apply(o *runOptions)
}

type runOptionFunc func(*runOptions)

func (f runOptionFunc) apply(o *runOptions) { f(o) }

func WithStdin(stdin io.Reader) RunOption {
	return runOptionFunc(func(o *runOptions) {
		o.stdin = stdin
	})
}

func WithStdout(stdout io.Writer) RunOption {
	return runOptionFunc(func(o *runOptions) {
		o.stdout = stdout
	})
}

func WithDir(dir string) RunOption {
	return runOptionFunc(func(o *runOptions) {
		o.dir = dir
	})
}
