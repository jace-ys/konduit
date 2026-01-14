package konduit_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jace-ys/konduit/pkg/konduit"
)

func TestInstance_New(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		args   []string
		values []string
		opts   []konduit.Option
		want   *konduit.Instance
	}{
		// Args
		{
			name: "parses args",
			args: []string{"install", "my-release", "my-chart"},
			want: &konduit.Instance{
				HelmArgs: []string{"install", "my-release", "my-chart"},
			},
		},
		// Values from parameter
		{
			name:   "extracts values from param",
			args:   []string{"install", "my-release"},
			values: []string{"values-1.yaml", "values-2.yaml"},
			want: &konduit.Instance{
				HelmArgs: []string{"install", "my-release"},
				Values:   []string{"values-1.yaml", "values-2.yaml"},
			},
		},
		{
			name:   "orders param values before arg values",
			args:   []string{"install", "-f", "args.yaml", "my-release"},
			values: []string{"param.yaml"},
			want: &konduit.Instance{
				HelmArgs: []string{"install", "my-release"},
				Values:   []string{"param.yaml", "args.yaml"},
			},
		},
		// Values from args
		{
			name: "extracts -f and --values with space",
			args: []string{"install", "-f", "values-1.yaml", "--values", "values-2.yaml", "my-release"},
			want: &konduit.Instance{
				HelmArgs: []string{"install", "my-release"},
				Values:   []string{"values-1.yaml", "values-2.yaml"},
			},
		},
		{
			name: "extracts -f= and --values= inline",
			args: []string{"install", "-f=values-1.yaml", "--values=values-2.yaml", "my-release"},
			want: &konduit.Instance{
				HelmArgs: []string{"install", "my-release"},
				Values:   []string{"values-1.yaml", "values-2.yaml"},
			},
		},
		{
			name: "passes through orphan -f",
			args: []string{"install", "my-release", "-f"},
			want: &konduit.Instance{
				HelmArgs: []string{"install", "my-release", "-f"},
			},
		},
		// Patches from option
		{
			name: "extracts patches from option",
			args: []string{"install", "my-release"},
			opts: []konduit.Option{konduit.WithPatches([]string{"patches-1.yaml", "patches-2.yaml"})},
			want: &konduit.Instance{
				HelmArgs: []string{"install", "my-release"},
				Patches:  []string{"patches-1.yaml", "patches-2.yaml"},
			},
		},
		// Post-renderer from args
		{
			name: "extracts --post-renderer with space",
			args: []string{"install", "--post-renderer", "./bin/renderer", "my-release"},
			want: &konduit.Instance{
				HelmArgs:     []string{"install", "my-release"},
				PostRenderer: "./bin/renderer",
			},
		},
		{
			name: "extracts --post-renderer= inline",
			args: []string{"install", "--post-renderer=./bin/renderer", "my-release"},
			want: &konduit.Instance{
				HelmArgs:     []string{"install", "my-release"},
				PostRenderer: "./bin/renderer",
			},
		},
		{
			name: "passes through orphan --post-renderer",
			args: []string{"install", "my-release", "--post-renderer"},
			want: &konduit.Instance{
				HelmArgs: []string{"install", "my-release", "--post-renderer"},
			},
		},
		{
			name: "extracts --post-renderer-args mixed",
			args: []string{
				"install", "--post-renderer", "./bin/renderer", "my-release",
				"--post-renderer-args=--flag", "--post-renderer-args", "value",
			},
			want: &konduit.Instance{
				HelmArgs:         []string{"install", "my-release"},
				PostRenderer:     "./bin/renderer",
				PostRendererArgs: []string{"--flag", "value"},
			},
		},
		{
			name: "passes through orphan --post-renderer-args",
			args: []string{"install", "my-release", "--post-renderer-args"},
			want: &konduit.Instance{
				HelmArgs: []string{"install", "my-release", "--post-renderer-args"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := konduit.New(tt.args, tt.values, tt.opts...)
			require.NoError(t, err)

			assert.Equal(t, konduit.DefaultHelmCommand, actual.HelmCommand)
			assert.Equal(t, tt.want.HelmArgs, actual.HelmArgs)
			assert.Equal(t, tt.want.PostRenderer, actual.PostRenderer)
			assert.Equal(t, tt.want.PostRendererArgs, actual.PostRendererArgs)
			assert.Equal(t, tt.want.Values, actual.Values)
			assert.Equal(t, tt.want.Patches, actual.Patches)
		})
	}
}

func TestInstance_New_WithCUEEvaluator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		args   []string
		values []string
		opts   []konduit.Option
		want   *konduit.Instance
	}{
		{
			name:   "correctly separates values to be evaluated",
			args:   []string{"install", "my-release", "my-chart"},
			values: []string{"values.cue", "values.yaml"},
			want: &konduit.Instance{
				HelmArgs:         []string{"install", "my-release", "my-chart"},
				Values:           []string{"values.yaml"},
				ValuesToEvaluate: []string{"values.cue"},
			},
		},
		{
			name: "correctly separates patches to be evaluated",
			args: []string{"install", "my-release", "my-chart"},
			opts: []konduit.Option{konduit.WithPatches([]string{"patches.cue", "patches.yaml"})},
			want: &konduit.Instance{
				HelmArgs:          []string{"install", "my-release", "my-chart"},
				Patches:           []string{"patches.yaml"},
				PatchesToEvaluate: []string{"patches.cue"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			eval := konduit.NewCUEEvaluator()
			tt.opts = append(tt.opts, konduit.WithEvaluator(eval))

			actual, err := konduit.New(tt.args, tt.values, tt.opts...)
			require.NoError(t, err)

			assert.Equal(t, konduit.DefaultHelmCommand, actual.HelmCommand)
			assert.Equal(t, tt.want.HelmArgs, actual.HelmArgs)
			assert.Equal(t, tt.want.Values, actual.Values)
			assert.Equal(t, tt.want.ValuesToEvaluate, actual.ValuesToEvaluate)
			assert.Equal(t, tt.want.Patches, actual.Patches)
			assert.Equal(t, tt.want.PatchesToEvaluate, actual.PatchesToEvaluate)
		})
	}
}
