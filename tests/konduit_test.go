package integration_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest
func TestKonduitCUE(t *testing.T) {
	gomega.RegisterTestingT(t)

	konduit, err := gexec.Build("github.com/jace-ys/konduit/cmd/konduit")
	require.NoError(t, err)
	t.Cleanup(gexec.CleanupBuildArtifacts)

	secrets, err := json.Marshal(map[string]any{
		"secrets": map[string]any{
			"ELASTICSEARCH_MONITORING_ES_PASSWORD": "secret123",
		},
	})
	require.NoError(t, err)

	for _, environment := range []string{"development", "production"} {
		t.Run(environment, func(t *testing.T) {
			cmd := exec.CommandContext(t.Context(), konduit, "cue",
				"--cue-base-dir", "../examples",
				"-v", fmt.Sprintf("../examples/logstash/values/%s/values.cue", environment),
				"-v", fmt.Sprintf("../examples/logstash/values/%s/values.yaml", environment),
				"-p", "../examples/logstash/patches/patches.cue",
				"-p", "../examples/logstash/patches/patches.yaml",
				"-s", string(secrets),
				"-s", fmt.Sprintf("@../examples/data/%s.json", environment),
				"--",
				"template", "logstash", "../examples/logstash/eck-logstash-0.17.0.tgz",
				"--post-renderer", "../hack/post-render-chain",
				"--post-renderer-args", "kustomize-1",
				"--post-renderer-args", "kustomize-2",
			)

			var stdout bytes.Buffer
			var stderr bytes.Buffer

			session, err := gexec.Start(cmd, &stdout, &stderr)
			require.NoError(t, err)
			session.Wait(10 * time.Second)
			require.Equal(t, 0, session.ExitCode(), stderr.String())

			testdata, err := os.ReadFile(fmt.Sprintf("testdata/logstash-%s.yaml", environment))
			require.NoError(t, err)

			want := decodeYAMLDocs(t, testdata)
			actual := decodeYAMLDocs(t, stdout.Bytes())

			require.Len(t, actual, len(want))
			for i := range want {
				assert.Equal(t, want[i], actual[i])
			}
		})
	}
}

func decodeYAMLDocs(t *testing.T, data []byte) []map[string]any {
	t.Helper()

	var docs []map[string]any
	decoder := yaml.NewDecoder(bytes.NewReader(data))

	for {
		doc := make(map[string]any)
		if err := decoder.Decode(&doc); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			require.NoError(t, err)
		}
		docs = append(docs, doc)
	}

	return docs
}
