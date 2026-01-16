package konduit

import (
	"fmt"

	"cuelang.org/go/encoding/yaml"

	"github.com/jace-ys/konduit/pkg/cueval"
)

//mockery:generate: true
type Evaluator interface {
	Evaluate(files []string) (result []byte, err error)
	SupportedFileExt() string
}

type Evaluation struct {
	Files      []string `json:"files,omitempty"`
	ResultYAML string   `json:"result,omitempty"`
}

type NoopEvaluator struct{}

func NewNoopEvaluator() *NoopEvaluator {
	return &NoopEvaluator{}
}

func (e *NoopEvaluator) Evaluate(files []string) ([]byte, error) {
	return []byte{}, nil
}

func (e *NoopEvaluator) SupportedFileExt() string {
	return ""
}

type CUEEvaluator struct {
	opts []cueval.Option
}

func NewCUEEvaluator(opts ...cueval.Option) *CUEEvaluator {
	return &CUEEvaluator{opts: opts}
}

func (e *CUEEvaluator) Evaluate(files []string) ([]byte, error) {
	value, err := cueval.Eval(files, e.opts...)
	if err != nil {
		return nil, fmt.Errorf("evaluate CUE: %w", err)
	}

	result, err := yaml.Encode(value)
	if err != nil {
		return nil, fmt.Errorf("encode CUE value: %w", err)
	}

	return result, nil
}

func (e *CUEEvaluator) SupportedFileExt() string {
	return ".cue"
}
