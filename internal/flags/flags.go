package flags

import (
	"flag"
)

type Flags struct {
	File          string
	EnvVars       FlagStringSlice
	SshConfig     string
	OnlyHosts     string
	ExceptHosts   string
	Debug         bool
	DisablePrefix bool
	ShowVersion   bool
	ShowHelp      bool
}

func New() *Flags {
	var f Flags

	flag.StringVar(&f.File, "f", "", "Custom path to ./Supfile[.yml]")
	flag.Var(&f.EnvVars, "e", "Set environment variables")
	flag.Var(&f.EnvVars, "env", "Set environment variables")
	flag.StringVar(&f.SshConfig, "sshconfig", "", "Read SSH Config file, ie. ~/.ssh/config file")
	flag.StringVar(&f.OnlyHosts, "only", "", "Filter hosts using regexp")
	flag.StringVar(&f.ExceptHosts, "except", "", "Filter out hosts using regexp")
	flag.BoolVar(&f.Debug, "D", false, "Enable debug mode")
	flag.BoolVar(&f.Debug, "debug", false, "Enable debug mode")
	flag.BoolVar(&f.DisablePrefix, "disable-prefix", false, "Disable hostname prefix")
	flag.BoolVar(&f.ShowVersion, "v", false, "Print version")
	flag.BoolVar(&f.ShowVersion, "version", false, "Print version")
	flag.BoolVar(&f.ShowHelp, "h", false, "Show help")
	flag.BoolVar(&f.ShowHelp, "help", false, "Show help")

	flag.Parse()

	return &f
}

// Wrapper for flag Args function
func Args() []string {
	return flag.Args()
}

// Wrapper for flag PrintDefaults function
func PrintDefaults() {
	flag.PrintDefaults()
}
