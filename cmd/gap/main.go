package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"sail/sail-golang-game-base/go-agent-platform-cli/internal/scaffold"
)

const version = "0.5.0"

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "project", "new":
		projectCmd(os.Args[2:])
	case "module", "gen":
		moduleCmd(os.Args[2:])
	case "ctrl":
		ctrlCmd(os.Args[2:])
	case "service":
		serviceCmd(os.Args[2:])
	case "dao":
		daoCmd(os.Args[2:])
	case "version", "-v", "--version":
		fmt.Println(version)
	case "help", "-h", "--help":
		usage()
	default:
		fail(fmt.Errorf("unknown command %q", os.Args[1]))
	}
}

func projectCmd(args []string) {
	spec, err := parseProjectSpec(args)
	if err != nil {
		fail(err)
	}

	result, err := scaffold.GenerateProject(spec)
	if err != nil {
		fail(err)
	}
	printResult("project", result)
}

func parseProjectSpec(args []string) (scaffold.ProjectSpec, error) {
	fs := flag.NewFlagSet("project", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	name := fs.String("name", "go-agent-platform", "project name")
	modulePath := fs.String("module", "", "go module path, for example: example.com/go-agent-platform")
	dir := fs.String("dir", "", "target directory")
	force := fs.Bool("force", false, "overwrite existing files")
	flagArgs, positionalArgs := splitProjectArgs(args)
	if err := fs.Parse(flagArgs); err != nil {
		return scaffold.ProjectSpec{}, err
	}
	if len(positionalArgs) > 1 {
		return scaffold.ProjectSpec{}, fmt.Errorf("project accepts only one project name, got %q", positionalArgs)
	}

	nameSet := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "name" {
			nameSet = true
		}
	})
	if len(positionalArgs) == 1 {
		if nameSet {
			return scaffold.ProjectSpec{}, fmt.Errorf("project name must use either the positional argument or -name, not both")
		}
		*name = positionalArgs[0]
	}

	return scaffold.ProjectSpec{
		Name:   *name,
		Module: *modulePath,
		Dir:    *dir,
		Force:  *force,
	}, nil
}

func splitProjectArgs(args []string) ([]string, []string) {
	valueFlags := map[string]bool{
		"-name": true, "--name": true,
		"-module": true, "--module": true,
		"-dir": true, "--dir": true,
	}
	flagArgs := make([]string, 0, len(args))
	positionals := make([]string, 0, 1)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if valueFlags[arg] {
			flagArgs = append(flagArgs, arg)
			if i+1 < len(args) {
				i++
				flagArgs = append(flagArgs, args[i])
			}
			continue
		}
		if strings.HasPrefix(arg, "-") {
			flagArgs = append(flagArgs, arg)
			continue
		}
		positionals = append(positionals, arg)
	}
	return flagArgs, positionals
}

func moduleCmd(args []string) {
	opts := parseModuleFlags("module", args)
	result, err := scaffold.GenerateModule(scaffold.ModuleSpec{
		Name:   opts.name,
		Root:   opts.root,
		Module: opts.modulePath,
		Force:  opts.force,
	})
	if err != nil {
		fail(err)
	}
	printResult("module", result)
}

func ctrlCmd(args []string) {
	opts := parseModuleFlags("ctrl", args)
	result, err := scaffold.GenerateCtrl(scaffold.CtrlSpec{
		Name:   opts.name,
		Root:   opts.root,
		Module: opts.modulePath,
		Force:  opts.force,
	})
	if err != nil {
		fail(err)
	}
	printResult("ctrl", result)
}

func serviceCmd(args []string) {
	opts := parseModuleFlags("service", args)
	result, err := scaffold.GenerateService(scaffold.ServiceSpec{
		Name:   opts.name,
		Root:   opts.root,
		Module: opts.modulePath,
		Force:  opts.force,
	})
	if err != nil {
		fail(err)
	}
	printResult("service", result)
}

func daoCmd(args []string) {
	opts := parseModuleFlags("dao", args)
	result, err := scaffold.GenerateDao(scaffold.DaoSpec{
		Name:   opts.name,
		Root:   opts.root,
		Module: opts.modulePath,
		Force:  opts.force,
	})
	if err != nil {
		fail(err)
	}
	printResult("dao", result)
}

type moduleOptions struct {
	name       string
	root       string
	modulePath string
	force      bool
}

func parseModuleFlags(command string, args []string) moduleOptions {
	fs := flag.NewFlagSet(command, flag.ExitOnError)
	name := fs.String("name", "", "module name, for example: orders")
	root := fs.String("root", ".", "platform project root")
	modulePath := fs.String("module", "", "override generated import module path")
	_ = fs.String("app", "", "deprecated: standalone projects generate directly under root")
	force := fs.Bool("force", false, "overwrite existing files")
	_ = fs.Parse(args)
	return moduleOptions{name: *name, root: *root, modulePath: *modulePath, force: *force}
}

func printResult(kind string, result scaffold.Result) {
	fmt.Printf("generated %s at %s\n", kind, result.Root)
	fmt.Printf("written: %d, skipped: %d\n", result.WrittenCount(), result.SkippedCount())
	for _, file := range result.Files {
		fmt.Printf("  %-7s %s\n", file.Status, file.Path)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "gap:", err)
	os.Exit(1)
}

func usage() {
	fmt.Print(`gap creates Go advanced web + Eino agent platform projects.

Usage:
  gap project gap-test
  gap project gap-test -module example.com/gap-test -dir ./gap-test
  gap project -name go-agent-platform -module example.com/go-agent-platform -dir ./go-agent-platform
  gap module  -name orders -root ./go-agent-platform
  gap ctrl    [-name orders] -root ./go-agent-platform
  gap service [-name orders] -root ./go-agent-platform
  gap dao     [-name orders] -root ./go-agent-platform

Commands:
  project      create a full platform project scaffold
  module       create root api/controller/service/logic/model/dao files
  ctrl         regenerate api interface + controllers from v1 DTOs (force all when -name omitted)
  service      regenerate model + service interface from v1 DTOs (force all when -name omitted)
  dao          regenerate dao + logic stubs from v1 DTOs (force all when -name omitted)
  version      print CLI version

Flags:
  -force       overwrite existing files
`)
	os.Exit(0)
}
