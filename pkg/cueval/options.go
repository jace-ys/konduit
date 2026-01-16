package cueval

import (
	"cuelang.org/go/cue/load"
)

const DefaultScopePath = "#Konduit"

type Evaluator struct {
	loader *load.Config
	scope  string
	scopes []string
}

func NewEvaluator(opts ...Option) *Evaluator {
	e := &Evaluator{
		loader: &load.Config{},
		scope:  DefaultScopePath,
	}

	for _, opt := range opts {
		opt.apply(e)
	}

	return e
}

type Option interface {
	apply(i *Evaluator)
}

type OptionFunc func(*Evaluator)

func (f OptionFunc) apply(o *Evaluator) { f(o) }

func WithLoadDir(dir string) Option {
	return OptionFunc(func(o *Evaluator) {
		o.loader.Dir = dir
	})
}

func WithLoadModuleRoot(root string) Option {
	return OptionFunc(func(o *Evaluator) {
		o.loader.ModuleRoot = root
	})
}

func WithScopePath(path string) Option {
	return OptionFunc(func(o *Evaluator) {
		o.scope = path
	})
}

func WithScopes(scopes ...string) Option {
	return OptionFunc(func(o *Evaluator) {
		o.scopes = append(o.scopes, scopes...)
	})
}
