package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/DTreshy/sup/internal/command"
	"github.com/DTreshy/sup/internal/envs"
	"github.com/DTreshy/sup/internal/network"
	"github.com/DTreshy/sup/internal/sup"
	"github.com/DTreshy/sup/internal/supfile"
	"github.com/mikkeloscar/sshconfig"
	"github.com/pkg/errors"
)

var (
	file          string
	envVars       flagStringSlice
	sshConfig     string
	onlyHosts     string
	exceptHosts   string
	debug         bool
	disablePrefix bool
	showVersion   bool
	showHelp      bool

	ErrUsage            = errors.New("Usage: sup [OPTIONS] NETWORK COMMAND [...]\n       sup [ --help | -v | --version ]")
	ErrUnknownNetwork   = errors.New("Unknown network")
	ErrNetworkNoHosts   = errors.New("No hosts defined for a given network")
	ErrCmd              = errors.New("Unknown command/target")
	ErrTargetNoCommands = errors.New("No commands defined for a given target")
	ErrConfigFile       = errors.New("Unknown ssh_config file")
)

type flagStringSlice []string

func (f *flagStringSlice) String() string {
	return fmt.Sprintf("%v", *f)
}

func (f *flagStringSlice) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func init() {
	flag.StringVar(&file, "f", "", "Custom path to ./Supfile[.yml]")
	flag.Var(&envVars, "e", "Set environment variables")
	flag.Var(&envVars, "env", "Set environment variables")
	flag.StringVar(&sshConfig, "sshconfig", "", "Read SSH Config file, ie. ~/.ssh/config file")
	flag.StringVar(&onlyHosts, "only", "", "Filter hosts using regexp")
	flag.StringVar(&exceptHosts, "except", "", "Filter out hosts using regexp")
	flag.BoolVar(&debug, "D", false, "Enable debug mode")
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
	flag.BoolVar(&disablePrefix, "disable-prefix", false, "Disable hostname prefix")
	flag.BoolVar(&showVersion, "v", false, "Print version")
	flag.BoolVar(&showVersion, "version", false, "Print version")
	flag.BoolVar(&showHelp, "h", false, "Show help")
	flag.BoolVar(&showHelp, "help", false, "Show help")
}

func networkUsage(conf *supfile.Supfile) {
	w := &tabwriter.Writer{}

	w.Init(os.Stderr, 4, 4, 2, ' ', 0)

	defer w.Flush()

	// Print available networks/hosts.
	fmt.Fprintln(w, "Networks:\t")

	for _, name := range conf.Networks.Names {
		fmt.Fprintf(w, "- %v\n", name)

		net, _ := conf.Networks.Get(name)

		for _, host := range net.Hosts {
			fmt.Fprintf(w, "\t- %v\n", host)
		}
	}

	fmt.Fprintln(w)
}

func cmdUsage(conf *supfile.Supfile) {
	w := &tabwriter.Writer{}

	w.Init(os.Stderr, 4, 4, 2, ' ', 0)

	defer w.Flush()

	// Print available targets/commands.
	fmt.Fprintln(w, "Targets:\t")

	for _, name := range conf.Targets.Names {
		cmds, _ := conf.Targets.Get(name)
		fmt.Fprintf(w, "- %v\t%v\n", name, strings.Join(cmds, " "))
	}

	fmt.Fprintln(w, "\t")
	fmt.Fprintln(w, "Commands:\t")

	for _, name := range conf.Commands.Names {
		cmd, _ := conf.Commands.Get(name)
		fmt.Fprintf(w, "- %v\t%v\n", name, cmd.Desc)
	}

	fmt.Fprintln(w)
}

// parseArgs parses args and returns network and commands to be run.
// On error, it prints usage and exits.
func parseArgs(conf *supfile.Supfile) (*network.Network, []*command.Command, error) {
	var commands []*command.Command

	args := flag.Args()
	if len(args) < 1 {
		networkUsage(conf)
		return nil, nil, ErrUsage
	}

	// Does the <network> exist?
	net, ok := conf.Networks.Get(args[0])
	if !ok {
		networkUsage(conf)
		return nil, nil, ErrUnknownNetwork
	}

	// Parse CLI --env flag env vars, override values defined in Network env.
	for _, env := range envVars {
		if env == "" {
			continue
		}

		i := strings.Index(env, "=")
		if i < 0 {
			if len(env) > 0 {
				net.Env.Set(env, "")
			}

			continue
		}

		net.Env.Set(env[:i], env[i+1:])
	}

	hosts, err := net.ParseInventory()
	if err != nil {
		return nil, nil, err
	}

	net.Hosts = append(net.Hosts, hosts...)

	// Does the <network> have at least one host?
	if len(net.Hosts) == 0 {
		networkUsage(conf)
		return nil, nil, ErrNetworkNoHosts
	}

	// Check for the second argument
	if len(args) < 2 {
		cmdUsage(conf)
		return nil, nil, ErrUsage
	}

	// In case of the network.Env needs an initialization
	if net.Env == nil {
		net.Env = make(envs.EnvList, 0)
	}

	// Add default env variable with current network
	net.Env.Set("SUP_NETWORK", args[0])
	// Add default nonce
	net.Env.Set("SUP_TIME", time.Now().UTC().Format(time.RFC3339))

	if os.Getenv("SUP_TIME") != "" {
		net.Env.Set("SUP_TIME", os.Getenv("SUP_TIME"))
	}

	// Add user
	if os.Getenv("SUP_USER") != "" {
		net.Env.Set("SUP_USER", os.Getenv("SUP_USER"))
	} else {
		net.Env.Set("SUP_USER", os.Getenv("USER"))
	}

	for _, name := range args[1:] {
		// Target?
		target, isTarget := conf.Targets.Get(name)
		if isTarget {
			// Loop over target's commands.
			for _, cmdName := range target {
				cmd, isCommand := conf.Commands.Get(cmdName)
				if !isCommand {
					cmdUsage(conf)
					return nil, nil, fmt.Errorf("%v: %v", ErrCmd, cmdName)
				}

				cmd.Name = cmdName
				commands = append(commands, &cmd)
			}
		}

		// Command?
		cmd, isCommand := conf.Commands.Get(name)
		if isCommand {
			cmd.Name = name
			commands = append(commands, &cmd)
		}

		if !isTarget && !isCommand {
			cmdUsage(conf)
			return nil, nil, fmt.Errorf("%v: %v", ErrCmd, name)
		}
	}

	return &net, commands, nil
}

func resolvePath(path string) string {
	if path == "" {
		return ""
	}

	if path[:2] == "~/" {
		usr, err := user.Current()
		if err == nil {
			path = filepath.Join(usr.HomeDir, path[2:])
		}
	}

	return path
}

func main() {
	flag.Parse()

	if showHelp {
		fmt.Fprintln(os.Stderr, ErrUsage, "\n\nOptions:")
		flag.PrintDefaults()

		return
	}

	if showVersion {
		fmt.Fprintln(os.Stderr, sup.VERSION)
		return
	}

	if file == "" {
		file = "./Supfile"
	}

	data, err := os.ReadFile(resolvePath(file))
	if err != nil {
		firstErr := err

		data, err = os.ReadFile("./Supfile.yml") // Alternative to ./Supfile.
		if err != nil {
			fmt.Fprintln(os.Stderr, firstErr)
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	conf, err := supfile.NewSupfile(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Parse network and commands to be run from args.
	net, commands, err := parseArgs(conf)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// --only flag filters hosts
	if onlyHosts != "" {
		expr, err := regexp.CompilePOSIX(onlyHosts)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		var hosts []string

		for _, host := range net.Hosts {
			if expr.MatchString(host) {
				hosts = append(hosts, host)
			}
		}

		if len(hosts) == 0 {
			fmt.Fprintln(os.Stderr, fmt.Errorf("no hosts match --only '%v' regexp", onlyHosts))
			os.Exit(1)
		}

		net.Hosts = hosts
	}

	// --except flag filters out hosts
	if exceptHosts != "" {
		expr, err := regexp.CompilePOSIX(exceptHosts)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		var hosts []string

		for _, host := range net.Hosts {
			if !expr.MatchString(host) {
				hosts = append(hosts, host)
			}
		}

		if len(hosts) == 0 {
			fmt.Fprintln(os.Stderr, fmt.Errorf("no hosts left after --except '%v' regexp", onlyHosts))
			os.Exit(1)
		}

		net.Hosts = hosts
	}

	// --sshconfig flag location for ssh_config file
	if sshConfig != "" {
		confHosts, err := sshconfig.ParseSSHConfig(resolvePath(sshConfig))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// flatten Host -> *SSHHost, not the prettiest
		// but will do
		confMap := map[string]*sshconfig.SSHHost{}

		for _, conf := range confHosts {
			for _, host := range conf.Host {
				confMap[host] = conf
			}
		}

		// check network.Hosts for match
		for _, host := range net.Hosts {
			conf, found := confMap[host]
			if found {
				net.User = conf.User
				net.IdentityFile = resolvePath(conf.IdentityFile)
				net.Hosts = []string{fmt.Sprintf("%s:%d", conf.HostName, conf.Port)}
			}
		}
	}

	var vars envs.EnvList

	for _, val := range append(conf.Env, net.Env...) {
		vars.Set(val.Key, val.Value)
	}

	if err := vars.ResolveValues(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Parse CLI --env flag env vars, define $SUP_ENV and override values defined in Supfile.
	var cliVars envs.EnvList

	for _, env := range envVars {
		if env == "" {
			continue
		}

		i := strings.Index(env, "=")
		if i < 0 {
			if len(env) > 0 {
				vars.Set(env, "")
			}

			continue
		}

		vars.Set(env[:i], env[i+1:])
		cliVars.Set(env[:i], env[i+1:])
	}

	// SUP_ENV is generated only from CLI env vars.
	// Separate loop to omit duplicates.
	supEnv := ""

	for _, v := range cliVars {
		supEnv += fmt.Sprintf(" -e %v=%q", v.Key, v.Value)
	}

	vars.Set("SUP_ENV", strings.TrimSpace(supEnv))

	// Create new Stackup app.
	app, err := sup.New(conf)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	app.Debug(debug)
	app.Prefix(!disablePrefix)

	// Run all the commands in the given network.
	err = app.Run(net, vars, commands...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
