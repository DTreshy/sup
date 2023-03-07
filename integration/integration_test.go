package integration_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIntegration(t *testing.T) {
	dat, err := os.ReadFile("./commands.json")
	require.NoError(t, err)

	var scripts map[string][]string

	err = json.Unmarshal(dat, &scripts)
	require.NoError(t, err)

	for _, args := range scripts {
		command := exec.Command("./../bin/sup", args...)

		out, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("%s\n%s\n", string(out), err.Error())
		}
	}
}
