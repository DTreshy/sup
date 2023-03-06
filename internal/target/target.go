package target

import (
	"fmt"

	"github.com/DTreshy/sup/pkg/unmarshaller"
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

	items, err := unmarshaller.Unmarshal(unmarshal)
	if err != nil {
		return fmt.Errorf("cannot parse targets: %w", err)
	}

	t.Names = make([]string, len(items))
	i := 0

	for key := range items {
		t.Names[i] = key
		i += 1
	}

	return nil
}

func (t *Targets) Get(name string) ([]string, bool) {
	cmds, ok := t.targets[name]
	return cmds, ok
}
