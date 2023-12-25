package network

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/DTreshy/sup/internal/envs"
	"github.com/DTreshy/sup/internal/flags"
)

// Network is group of hosts with extra custom env vars.
type Network struct {
	Env       envs.EnvList `yaml:"env"`
	Inventory string       `yaml:"inventory"`
	Hosts     []string     `yaml:"hosts"`
	Bastion   string       `yaml:"bastion"` // Jump host for the environment

	// Should these live on Hosts too? We'd have to change []string to struct, even in Supfile.
	User         string // `yaml:"user"`
	IdentityFile string // `yaml:"identity_file"`
}

// ParseInventory runs the inventory command, if provided, and appends
// the command's output lines to the manually defined list of hosts.
func (n Network) ParseInventory(envs *envs.EnvList) ([]string, error) {
	if n.Inventory == "" {
		return nil, nil
	}

	parseInventoryArgs := []string{
		"-c",
		n.Inventory,
	}
	cmd := exec.Command("/bin/sh", parseInventoryArgs...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, envs.Slice()...)
	cmd.Env = append(cmd.Env, n.Env.Slice()...)
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var hosts []string

	buf := bytes.NewBuffer(output)

	for {
		host, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		host = strings.TrimSpace(host)
		// skip empty lines and comments
		if host == "" || host[:1] == "#" {
			continue
		}

		hosts = append(hosts, host)
	}

	return hosts, nil
}

func (n *Network) SetEnvs(vars flags.FlagStringSlice) {
	// Parse CLI --env flag env vars, override values defined in Network env.
	for _, env := range vars {
		if env == "" {
			continue
		}

		val := strings.SplitN(env, "=", 2)
		if len(val) == 1 {
			n.Env.Set(env, "")

			continue
		}

		n.Env.Set(val[0], val[1])
	}
}
