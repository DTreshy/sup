package network

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/DTreshy/sup/internal/envs"
	"github.com/DTreshy/sup/pkg/yamlparser"
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
func (n Network) ParseInventory() ([]string, error) {
	if n.Inventory == "" {
		return nil, nil
	}

	parseInventoryArgs := []string{
		"-c",
		n.Inventory,
	}
	cmd := exec.Command("/bin/sh", parseInventoryArgs...)
	cmd.Env = os.Environ()
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

// Networks is a list of user-defined networks
type Networks struct {
	Names []string
	Nets  map[string]Network
}

func (n *Networks) UnmarshalYAML(unmarshal func(any) error) error {
	err := unmarshal(&n.Nets)
	if err != nil {
		return err
	}

	var ok bool

	items, err := yamlparser.Unmarshal(unmarshal)
	if err != nil {
		return fmt.Errorf("cannot parse networks: %w", err)
	}

	n.Names = make([]string, len(items))

	for i, item := range items {
		n.Names[i], ok = item.Key.(string)
		if !ok {
			return fmt.Errorf("assertion to string failed")
		}
	}

	return nil
}

func (n *Networks) Get(name string) (Network, bool) {
	net, ok := n.Nets[name]
	return net, ok
}
