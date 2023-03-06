package envs

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/DTreshy/sup/pkg/unmarshaller"
	"github.com/pkg/errors"
)

// EnvVar represents an environment variable
type EnvVar struct {
	Key   string
	Value string
}

func (e EnvVar) String() string {
	return e.Key + `=` + e.Value
}

// AsExport returns the environment variable as a bash export statement
func (e EnvVar) AsExport() string {
	return `export ` + e.Key + `="` + e.Value + `";`
}

// EnvList is a list of environment variables that maps to a YAML map,
// but maintains order, enabling late variables to reference early variables.
type EnvList []*EnvVar

func (e EnvList) Slice() []string {
	envs := make([]string, len(e))

	for i, env := range e {
		envs[i] = env.String()
	}

	return envs
}

func (e *EnvList) UnmarshalYAML(unmarshal func(any) error) error {
	items, err := unmarshaller.Unmarshal(unmarshal)
	if err != nil {
		return fmt.Errorf("cannot unmarshal envs: %w", err)
	}

	for key, val := range items {
		e.Set(key, fmt.Sprintf("%v", val))
	}

	return nil
}

// Set key to be equal value in this list.
func (e *EnvList) Set(key, value string) {
	for i, v := range *e {
		if v.Key == key {
			(*e)[i].Value = value
			return
		}
	}

	*e = append(*e, &EnvVar{
		Key:   key,
		Value: value,
	})
}

func (e *EnvList) ResolveValues() error {
	if len(*e) == 0 {
		return nil
	}

	exports := ""

	for i, v := range *e {
		exports += v.AsExport()

		resolveValuesArgs := []string{
			"-c",
			exports + "echo -n " + v.Value + ";",
		}
		cmd := exec.Command("bash", resolveValuesArgs...)

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		cmd.Dir = cwd

		resolvedValue, err := cmd.Output()
		if err != nil {
			return errors.Wrapf(err, "resolving env var %v failed", v.Key)
		}

		(*e)[i].Value = string(resolvedValue)
	}

	return nil
}

func (e *EnvList) AsExport() string {
	// Process all ENVs into a string of form
	// `export FOO="bar"; export BAR="baz";`.
	exports := ``

	for _, v := range *e {
		exports += v.AsExport() + " "
	}

	return exports
}
