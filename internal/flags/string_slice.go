package flags

import "fmt"

type FlagStringSlice []string

func (f *FlagStringSlice) String() string {
	return fmt.Sprintf("%v", *f)
}

func (f *FlagStringSlice) Set(value string) error {
	*f = append(*f, value)
	return nil
}
