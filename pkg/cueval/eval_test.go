package cueval_test

import (
	"testing"

	"cuelang.org/go/encoding/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jace-ys/konduit/pkg/cueval"
)

func TestEval(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		files    []string
		opts     []cueval.Option
		wantYAML string
		wantErr  string
	}{
		{
			name:     "evaluates simple CUE file",
			files:    []string{"testdata/simple.cue"},
			wantYAML: "foo: hello\nbar: 42\n",
		},
		{
			name:  "evaluates CUE with JSON scope",
			files: []string{"testdata/scope.cue"},
			opts: []cueval.Option{
				cueval.WithScopes(`{"foo": "one", "bar": "two"}`),
			},
			wantYAML: "foo: one\nbar: two\n",
		},
		{
			name:  "evaluates CUE with YAML scope file",
			files: []string{"testdata/scope.cue"},
			opts: []cueval.Option{
				cueval.WithScopes("@testdata/scope.yaml"),
			},
			wantYAML: "foo: one\nbar: two\n",
		},
		{
			name:  "merges multiple scopes",
			files: []string{"testdata/scope.cue"},
			opts: []cueval.Option{
				cueval.WithScopes(`{"foo": "one"}`, `{"bar": "two"}`),
			},
			wantYAML: "foo: one\nbar: two\n",
		},
		{
			name:    "returns error when no files provided",
			wantErr: "no CUE files provided",
		},
		{
			name:    "returns error when value not concrete",
			files:   []string{"testdata/incomplete.cue"},
			wantErr: "value not concrete",
		},
		{
			name:  "returns error when scope file not found",
			files: []string{"testdata/simple.cue"},
			opts: []cueval.Option{
				cueval.WithScopes("@testdata/nonexistent.yaml"),
			},
			wantErr: "read scope file",
		},
		{
			name:  "returns error when scope data is invalid YAML",
			files: []string{"testdata/simple.cue"},
			opts: []cueval.Option{
				cueval.WithScopes(`{invalid yaml`),
			},
			wantErr: "extract scope data",
		},
		{
			name:  "returns error when scopes conflict with each other",
			files: []string{"testdata/scope.cue"},
			opts: []cueval.Option{
				cueval.WithScopes(`{"foo": "one"}`, `{"foo": "two"}`),
			},
			wantErr: "unify scopes",
		},
		{
			name:  "returns error when scope type conflicts",
			files: []string{"testdata/constrained.cue"},
			opts: []cueval.Option{
				cueval.WithScopes(`{"foo": "not-an-int", "bar": "fixed"}`),
			},
			wantErr: "build instance: foo: conflicting values",
		},
		{
			name:  "returns error when scope value conflicts",
			files: []string{"testdata/constrained.cue"},
			opts: []cueval.Option{
				cueval.WithScopes(`{"foo": 42, "bar": "different"}`),
			},
			wantErr: "build instance: bar: conflicting values",
		},
		{
			name:  "returns error when scope conflicts with CUE definition",
			files: []string{"testdata/konduit.cue"},
			opts: []cueval.Option{
				cueval.WithScopes(`{"foo": "different"}`),
			},
			wantErr: "unify instance with scopes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			value, err := cueval.Eval(tt.files, tt.opts...)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			result, err := yaml.Encode(value)
			require.NoError(t, err)
			assert.Equal(t, tt.wantYAML, string(result))
		})
	}
}
