package cueval

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/encoding/yaml"
)

func Eval(files []string, opts ...Option) (cue.Value, error) {
	return NewEvaluator(opts...).Eval(files)
}

func (e *Evaluator) Eval(files []string) (cue.Value, error) {
	if len(files) == 0 {
		return cue.Value{}, errors.New("no CUE files provided")
	}

	inst, err := e.Load(files)
	if err != nil {
		return cue.Value{}, err
	}

	v, err := e.Build(inst)
	if err != nil {
		return cue.Value{}, err
	}

	return v, nil
}

func (e *Evaluator) Load(files []string) (*build.Instance, error) {
	resolved := e.tryResolvePaths(files)

	instances := load.Instances(resolved, e.loader)
	if len(instances) != 1 {
		return nil, fmt.Errorf("expected 1 instance, got %d", len(instances))
	}

	inst := instances[0]
	if inst.Err != nil {
		return nil, fmt.Errorf("load instance: %w", inst.Err)
	}

	return inst, nil
}

func (e *Evaluator) tryResolvePaths(files []string) []string {
	if e.loader.Dir == "" {
		return files
	}

	dir, err := filepath.Abs(e.loader.Dir)
	if err != nil {
		return files
	}

	resolved := make([]string, len(files))
	for i, file := range files {
		abs, err := filepath.Abs(file)
		if err != nil {
			resolved[i] = file
			continue
		}

		rel, err := filepath.Rel(dir, abs)
		if err != nil {
			resolved[i] = file
			continue
		}

		resolved[i] = rel
	}

	return resolved
}

func (e *Evaluator) Build(inst *build.Instance) (cue.Value, error) {
	ctx := cuecontext.New()

	vScopes, err := e.buildScopes(ctx)
	if err != nil {
		return cue.Value{}, err
	}

	v := ctx.BuildInstance(inst, cue.Scope(vScopes))
	if v.Err() != nil {
		return cue.Value{}, fmt.Errorf("build instance: %w", v.Err())
	}

	v = v.Unify(vScopes)
	if v.Err() != nil {
		return cue.Value{}, fmt.Errorf("unify instance with scopes: %w", v.Err())
	}

	if err := v.Validate(cue.Concrete(true)); err != nil {
		return cue.Value{}, fmt.Errorf("value not concrete: %w", err)
	}

	return v, nil
}

func (e *Evaluator) buildScopes(ctx *cue.Context) (cue.Value, error) {
	vAllScopes := ctx.CompileString("{}")

	for _, scope := range e.scopes {
		if scope == "" {
			continue
		}

		vScope, err := e.parseScope(ctx, scope)
		if err != nil {
			return cue.Value{}, err
		}

		vAllScopes = vAllScopes.Unify(vScope)
		if vAllScopes.Err() != nil {
			return cue.Value{}, fmt.Errorf("unify scopes: %w", vAllScopes.Err())
		}
	}

	return vAllScopes, nil
}

func (e *Evaluator) parseScope(ctx *cue.Context, scope string) (cue.Value, error) {
	var data []byte

	if filename, ok := strings.CutPrefix(scope, "@"); ok {
		scopeData, err := os.ReadFile(filename)
		if err != nil {
			return cue.Value{}, fmt.Errorf("read scope file: %w", err)
		}
		data = scopeData
	} else {
		data = []byte(scope)
	}

	ast, err := yaml.Extract("", data)
	if err != nil {
		return cue.Value{}, fmt.Errorf("extract scope data: %w", err)
	}

	vScope := ctx.CompileString("{}")

	vScope = vScope.FillPath(cue.ParsePath(e.scope), ast)
	if vScope.Err() != nil {
		return cue.Value{}, fmt.Errorf("populate scope data: %w", vScope.Err())
	}

	return vScope, nil
}
