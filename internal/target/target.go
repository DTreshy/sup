package target

import (
	"fmt"

	"github.com/DTreshy/sup/pkg/yamlparser"
)

// Targets is a list of user-defined targets
type Targets struct {
	Names   []string
	targets map[string][]string
}

func (t *Targets) UnmarshalYAML(unmarshal func(any) error) error {
	err := unmarshal(&t.targets)
	if err != nil {
		return err
	}

	var ok bool

	items, err := yamlparser.Unmarshal(unmarshal)
	if err != nil {
		return fmt.Errorf("cannot parse targets: %w", err)
	}

	t.Names = make([]string, len(items))
	for i, item := range items {
		t.Names[i], ok = item.Key.(string)
		if !ok {
			return fmt.Errorf("assertion to string failed")
		}
	}

	return nil
}

func (t *Targets) Get(name string) ([]string, bool) {
	cmds, ok := t.targets[name]
	return cmds, ok
}
