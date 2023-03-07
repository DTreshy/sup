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

	var m map[string][]string

	err = json.Unmarshal(dat, &m)
	require.NoError(t, err)

	for _, args := range m {
		command := exec.Command("./../bin/sup", args...)

		err = command.Run()
		require.NoError(t, err)
	}
}
