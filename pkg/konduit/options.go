package konduit

import "github.com/jace-ys/konduit/internal/exec"

type Option interface {
	Apply(i *Instance)
}

type OptionFunc func(*Instance)

func (o OptionFunc) Apply(i *Instance) { o(i) }

func WithHelmCommand(command string) Option {
	return OptionFunc(func(i *Instance) {
		i.HelmCommand = command
	})
}

func WithPatches(patches []string) Option {
	return OptionFunc(func(i *Instance) {
		i.patchesOpt = patches
	})
}

func WithWorkDir(dir string) Option {
	return OptionFunc(func(i *Instance) {
		i.dir = dir
	})
}

func WithEvaluator(evaluator Evaluator) Option {
	return OptionFunc(func(i *Instance) {
		i.evaluator = evaluator
	})
}

func WithRunner(runner exec.Runner) Option {
	return OptionFunc(func(i *Instance) {
		i.runner = runner
	})
}
