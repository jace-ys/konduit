package exec

import (
	"context"
	"io"
)

//mockery:generate: true
type Runner interface {
	Run(ctx context.Context, command string, args []string, opts ...Option) error
}

type Options struct {
	stdin  io.Reader
	stdout io.Writer
	dir    string
}

type Option interface {
	Apply(o *Options)
}

type OptionFunc func(*Options)

func (f OptionFunc) Apply(o *Options) { f(o) }

func WithStdin(stdin io.Reader) OptionFunc {
	return OptionFunc(func(o *Options) {
		o.stdin = stdin
	})
}

func WithStdout(stdout io.Writer) OptionFunc {
	return OptionFunc(func(o *Options) {
		o.stdout = stdout
	})
}

func WithDir(dir string) OptionFunc {
	return OptionFunc(func(o *Options) {
		o.dir = dir
	})
}
