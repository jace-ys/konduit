package konduit_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jace-ys/konduit/internal/kustomize"
	"github.com/jace-ys/konduit/pkg/konduit"
	"github.com/jace-ys/konduit/pkg/konduit/mocks"
)

func TestInstance_Construct(t *testing.T) {
	t.Parallel()

	executable, err := os.Executable()
	require.NoError(t, err)

	tests := []struct {
		name     string
		instance *konduit.Instance
		want     *konduit.Invocation
	}{
		// Basic
		{
			name: "passes through args",
			instance: &konduit.Instance{
				HelmArgs: []string{"template", "my-release", "my-chart"},
			},
			want: &konduit.Invocation{
				Args: []string{"template", "my-release", "my-chart"},
			},
		},
		// Values
		{
			name: "adds --values for each value",
			instance: &konduit.Instance{
				HelmArgs: []string{"template", "my-release"},
				Values:   []string{"values-1.yaml", "values-2.yaml"},
			},
			want: &konduit.Invocation{
				Args: []string{
					"template", "my-release",
					"--values", "values-1.yaml",
					"--values", "values-2.yaml",
				},
			},
		},
		{
			name: "adds evaluated.yaml for evaluated values",
			instance: &konduit.Instance{
				HelmArgs:         []string{"template", "my-release"},
				Values:           []string{"values.yaml"},
				ValuesToEvaluate: []string{"values.cue"},
			},
			want: &konduit.Invocation{
				Args: []string{
					"template", "my-release",
					"--values", "/tmp/evaluated.yaml",
					"--values", "values.yaml",
				},
			},
		},
		// Post-renderer without patches (passthrough)
		{
			name: "passes through post-renderer without patches",
			instance: &konduit.Instance{
				HelmArgs:     []string{"template", "my-release"},
				PostRenderer: "./bin/renderer",
			},
			want: &konduit.Invocation{
				Args: []string{
					"template", "my-release",
					"--post-renderer", "./bin/renderer",
				},
			},
		},
		{
			name: "passes through post-renderer-args without patches",
			instance: &konduit.Instance{
				HelmArgs:         []string{"template", "my-release"},
				PostRenderer:     "./bin/renderer",
				PostRendererArgs: []string{"--flag", "value"},
			},
			want: &konduit.Invocation{
				Args: []string{
					"template", "my-release",
					"--post-renderer", "./bin/renderer",
					"--post-renderer-args", "--flag",
					"--post-renderer-args", "value",
				},
			},
		},
		// Patches
		{
			name: "adds konduit post-renderer for patches",
			instance: &konduit.Instance{
				HelmArgs: []string{"template", "my-release", "my-chart"},
				Patches:  []string{"p1.yaml"},
			},
			want: &konduit.Invocation{
				Args: []string{
					"template", "my-release", "my-chart",
					"--post-renderer", executable,
					"--post-renderer-args", "kustomize",
					"--post-renderer-args", "--dir",
					"--post-renderer-args", "/tmp",
				},
			},
		},
		{
			name: "adds konduit post-renderer for evaluated patches",
			instance: &konduit.Instance{
				HelmArgs:          []string{"template", "my-release"},
				PatchesToEvaluate: []string{"patches.cue"},
			},
			want: &konduit.Invocation{
				Args: []string{
					"template", "my-release",
					"--post-renderer", executable,
					"--post-renderer-args", "kustomize",
					"--post-renderer-args", "--dir",
					"--post-renderer-args", "/tmp",
				},
			},
		},
		{
			name: "chains existing post-renderer through konduit",
			instance: &konduit.Instance{
				HelmArgs:     []string{"template", "my-release"},
				Patches:      []string{"patches.yaml"},
				PostRenderer: "./bin/renderer",
			},
			want: &konduit.Invocation{
				Args: []string{
					"template", "my-release",
					"--post-renderer", executable,
					"--post-renderer-args", "kustomize",
					"--post-renderer-args", "--dir",
					"--post-renderer-args", "/tmp",
					"--post-renderer-args", "--post-renderer",
					"--post-renderer-args", "./bin/renderer",
				},
			},
		},
		{
			name: "passes post-renderer-args through chain",
			instance: &konduit.Instance{
				HelmArgs:         []string{"template", "my-release"},
				Patches:          []string{"patches.yaml"},
				PostRenderer:     "./bin/renderer",
				PostRendererArgs: []string{"--flag", "value"},
			},
			want: &konduit.Invocation{
				Args: []string{
					"template", "my-release",
					"--post-renderer", executable,
					"--post-renderer-args", "kustomize",
					"--post-renderer-args", "--dir",
					"--post-renderer-args", "/tmp",
					"--post-renderer-args", "--post-renderer",
					"--post-renderer-args", "./bin/renderer",
					"--post-renderer-args", "--post-renderer-args",
					"--post-renderer-args", "--flag",
					"--post-renderer-args", "--post-renderer-args",
					"--post-renderer-args", "value",
				},
			},
		},
		// Combined
		{
			name: "combines values and patches",
			instance: &konduit.Instance{
				HelmArgs: []string{"template", "my-release"},
				Values:   []string{"values.yaml"},
				Patches:  []string{"patches.yaml"},
			},
			want: &konduit.Invocation{
				Args: []string{
					"template", "my-release",
					"--values", "values.yaml",
					"--post-renderer", executable,
					"--post-renderer-args", "kustomize",
					"--post-renderer-args", "--dir",
					"--post-renderer-args", "/tmp",
				},
			},
		},
		{
			name: "combines evaluated values and patches",
			instance: &konduit.Instance{
				HelmArgs:          []string{"template", "my-release"},
				Values:            []string{"values.yaml"},
				ValuesToEvaluate:  []string{"values.cue"},
				Patches:           []string{"patches.yaml"},
				PatchesToEvaluate: []string{"patches.cue"},
			},
			want: &konduit.Invocation{
				Args: []string{
					"template", "my-release",
					"--values", "/tmp/evaluated.yaml",
					"--values", "values.yaml",
					"--post-renderer", executable,
					"--post-renderer-args", "kustomize",
					"--post-renderer-args", "--dir",
					"--post-renderer-args", "/tmp",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.instance.HelmCommand = konduit.DefaultHelmCommand
			konduit.WithEvaluator(konduit.NewNoopEvaluator()).Apply(tt.instance)
			konduit.WithWorkDir("/tmp").Apply(tt.instance)

			actual, err := tt.instance.Construct()
			require.NoError(t, err)

			assert.Equal(t, konduit.DefaultHelmCommand, actual.Command)
			assert.Equal(t, tt.want.Args, actual.Args)
			assert.Equal(t, tt.instance.Values, actual.Values)
			assert.Equal(t, tt.instance.Patches, actual.Patches)
		})
	}
}

func TestInstance_Construct_WithEvaluator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		instance             *konduit.Instance
		setupMock            func(*mocks.MockEvaluator)
		wantEvaluatedValues  *konduit.Evaluation
		wantEvaluatedPatches *konduit.Evaluation
		wantErr              string
	}{
		{
			name: "returns evaluated values",
			instance: &konduit.Instance{
				ValuesToEvaluate: []string{"values.cue"},
			},
			setupMock: func(m *mocks.MockEvaluator) {
				m.EXPECT().Evaluate([]string{"values.cue"}).Return([]byte("key: value\n"), nil)
			},
			wantEvaluatedValues: &konduit.Evaluation{
				Files:      []string{"values.cue"},
				ResultYAML: "key: value\n",
			},
			wantEvaluatedPatches: &konduit.Evaluation{},
		},
		{
			name: "returns evaluated patches",
			instance: &konduit.Instance{
				PatchesToEvaluate: []string{"patches.cue"},
			},
			setupMock: func(m *mocks.MockEvaluator) {
				m.EXPECT().Evaluate([]string{"patches.cue"}).Return([]byte("key: value\n"), nil)
			},
			wantEvaluatedValues: &konduit.Evaluation{},
			wantEvaluatedPatches: &konduit.Evaluation{
				Files:      []string{"patches.cue"},
				ResultYAML: "key: value\n",
			},
		},
		{
			name: "returns evaluated values and patches",
			instance: &konduit.Instance{
				ValuesToEvaluate:  []string{"values.cue"},
				PatchesToEvaluate: []string{"patches.cue"},
			},
			setupMock: func(m *mocks.MockEvaluator) {
				m.EXPECT().Evaluate([]string{"values.cue"}).Return([]byte("key: value\n"), nil)
				m.EXPECT().Evaluate([]string{"patches.cue"}).Return([]byte("key: value\n"), nil)
			},
			wantEvaluatedValues: &konduit.Evaluation{
				Files:      []string{"values.cue"},
				ResultYAML: "key: value\n",
			},
			wantEvaluatedPatches: &konduit.Evaluation{
				Files:      []string{"patches.cue"},
				ResultYAML: "key: value\n",
			},
		},
		{
			name: "returns error when evaluating values fails",
			instance: &konduit.Instance{
				ValuesToEvaluate: []string{"invalid.cue"},
			},
			setupMock: func(m *mocks.MockEvaluator) {
				m.EXPECT().Evaluate([]string{"invalid.cue"}).Return(nil, assert.AnError)
			},
			wantErr: "evaluate values",
		},
		{
			name: "returns error when evaluating patches fails",
			instance: &konduit.Instance{
				PatchesToEvaluate: []string{"invalid.cue"},
			},
			setupMock: func(m *mocks.MockEvaluator) {
				m.EXPECT().Evaluate([]string{"invalid.cue"}).Return(nil, assert.AnError)
			},
			wantErr: "evaluate patches",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.instance.HelmCommand = konduit.DefaultHelmCommand
			tt.instance.HelmArgs = []string{"template", "my-release"}

			eval := mocks.NewMockEvaluator(t)
			tt.setupMock(eval)
			konduit.WithEvaluator(eval).Apply(tt.instance)

			actual, err := tt.instance.Construct()
			if tt.wantErr != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantEvaluatedValues, actual.EvaluatedValues)
			assert.Equal(t, tt.wantEvaluatedPatches, actual.EvaluatedPatches)
		})
	}
}

func TestInstance_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		instance           *konduit.Instance
		setupMockEvaluator func(*mocks.MockEvaluator)
		setupMockRunner    func(*mocks.MockRunner)
		wantYAML           map[string]string
		wantErr            string
	}{
		{
			name: "passes args to runner",
			instance: &konduit.Instance{
				HelmArgs: []string{"template", "my-release", "my-chart"},
			},
			setupMockEvaluator: func(m *mocks.MockEvaluator) {},
			setupMockRunner: func(m *mocks.MockRunner) {
				m.EXPECT().
					Run(mock.Anything, konduit.DefaultHelmCommand, []string{"template", "my-release", "my-chart"}).
					Return(nil)
			},
		},
		{
			name: "passes values to runner",
			instance: &konduit.Instance{
				HelmArgs: []string{"template", "my-release"},
				Values:   []string{"values.yaml"},
			},
			setupMockEvaluator: func(m *mocks.MockEvaluator) {},
			setupMockRunner: func(m *mocks.MockRunner) {
				m.EXPECT().
					Run(mock.Anything, konduit.DefaultHelmCommand, []string{"template", "my-release", "--values", "values.yaml"}).
					Return(nil)
			},
		},
		{
			name: "writes evaluated values to file",
			instance: &konduit.Instance{
				HelmArgs:         []string{"template", "my-release"},
				ValuesToEvaluate: []string{"values.cue"},
			},
			setupMockEvaluator: func(m *mocks.MockEvaluator) {
				m.EXPECT().Evaluate([]string{"values.cue"}).Return([]byte("key: value\n"), nil)
			},
			setupMockRunner: func(m *mocks.MockRunner) {
				m.EXPECT().Run(mock.Anything, konduit.DefaultHelmCommand, mock.Anything).Return(nil)
			},
			wantYAML: map[string]string{
				konduit.ValuesFile: "key: value\n",
			},
		},
		{
			name: "writes evaluated patches to file",
			instance: &konduit.Instance{
				HelmArgs:          []string{"template", "my-release"},
				PatchesToEvaluate: []string{"patches.cue"},
			},
			setupMockEvaluator: func(m *mocks.MockEvaluator) {
				m.EXPECT().Evaluate([]string{"patches.cue"}).Return([]byte("namePrefix: test-"), nil)
			},
			setupMockRunner: func(m *mocks.MockRunner) {
				m.EXPECT().Run(mock.Anything, konduit.DefaultHelmCommand, mock.Anything).Return(nil)
			},
			wantYAML: map[string]string{
				kustomize.KustomizationFile: `kind: Kustomization
apiVersion: kustomize.config.k8s.io/v1beta1
resources:
- manifests.yaml
namePrefix: test-
`,
			},
		},
		{
			name: "returns error when evaluator fails",
			instance: &konduit.Instance{
				HelmArgs:         []string{"template", "my-release"},
				ValuesToEvaluate: []string{"invalid.cue"},
			},
			setupMockEvaluator: func(m *mocks.MockEvaluator) {
				m.EXPECT().Evaluate([]string{"invalid.cue"}).Return(nil, assert.AnError)
			},
			setupMockRunner: func(m *mocks.MockRunner) {},
			wantErr:         "evaluate values",
		},
		{
			name: "returns error when runner fails",
			instance: &konduit.Instance{
				HelmArgs: []string{"template", "my-release"},
			},
			setupMockEvaluator: func(m *mocks.MockEvaluator) {},
			setupMockRunner: func(m *mocks.MockRunner) {
				m.EXPECT().Run(mock.Anything, konduit.DefaultHelmCommand, mock.Anything).Return(assert.AnError)
			},
			wantErr: "run invocation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()

			tt.instance.HelmCommand = konduit.DefaultHelmCommand
			konduit.WithWorkDir(dir).Apply(tt.instance)

			eval := mocks.NewMockEvaluator(t)
			tt.setupMockEvaluator(eval)
			konduit.WithEvaluator(eval).Apply(tt.instance)

			runner := mocks.NewMockRunner(t)
			tt.setupMockRunner(runner)
			konduit.WithRunner(runner).Apply(tt.instance)

			err := tt.instance.Execute(t.Context())
			if tt.wantErr != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			for filename, want := range tt.wantYAML {
				actual, err := os.ReadFile(filepath.Join(dir, filename))
				require.NoError(t, err)
				assert.YAMLEq(t, want, string(actual))
			}
		})
	}
}
