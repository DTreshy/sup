package network

import (
	"fmt"

	"github.com/DTreshy/sup/pkg/unmarshaller"
)

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

	items, err := unmarshaller.Unmarshal(unmarshal)
	if err != nil {
		return fmt.Errorf("cannot parse networks: %w", err)
	}

	n.Names = make([]string, len(items))
	i := 0

	for key := range items {
		n.Names[i] = key
		i += 1
	}

	return nil
}

func (n *Networks) Get(name string) (Network, bool) {
	net, ok := n.Nets[name]
	return net, ok
}
