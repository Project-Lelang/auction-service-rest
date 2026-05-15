package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/swaggo/swag/format"
	"github.com/swaggo/swag/gen"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: swag <init|fmt> [flags]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init", "i":
		runInit(os.Args[2:])
	case "fmt", "f":
		runFmt(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func runInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	searchDir := fs.String("d", "./", "Directories to parse, comma separated")
	generalInfo := fs.String("g", "main.go", "Go file with swagger general API info")
	output := fs.String("o", "./docs", "Output directory")
	outputTypes := fs.String("outputTypes", "go,json,yaml", "Output types: go,json,yaml")
	parseDep := fs.Bool("parseDependency", false, "Parse dependency folders")
	parseInternal := fs.Bool("parseInternal", false, "Parse internal packages")
	parseVendor := fs.Bool("parseVendor", false, "Parse vendor folder")
	parseDepth := fs.Int("parseDepth", 100, "Dependency parse depth")
	instanceName := fs.String("instanceName", "", "Swagger instance name")
	overridesFile := fs.String("overridesFile", gen.DefaultOverridesFile, "Global type overrides file")
	parseGoList := fs.Bool("parseGoList", true, "Parse dependency via go list")

	if err := fs.Parse(args); err != nil {
		log.Fatal(err)
	}

	types := strings.Split(*outputTypes, ",")

	cfg := &gen.Config{
		SearchDir:       *searchDir,
		MainAPIFile:     *generalInfo,
		OutputDir:       *output,
		OutputTypes:     types,
		ParseDependency: *parseDep,
		ParseInternal:   *parseInternal,
		ParseVendor:     *parseVendor,
		ParseDepth:      *parseDepth,
		InstanceName:    *instanceName,
		OverridesFile:   *overridesFile,
		ParseGoList:     *parseGoList,
		Debugger:        log.New(os.Stdout, "", log.LstdFlags),
	}

	if err := gen.New().Build(cfg); err != nil {
		log.Fatal(err)
	}
}

func runFmt(args []string) {
	fs := flag.NewFlagSet("fmt", flag.ExitOnError)
	searchDir := fs.String("d", "./", "Directories to parse, comma separated")
	generalInfo := fs.String("g", "main.go", "Go file with swagger general API info")

	if err := fs.Parse(args); err != nil {
		log.Fatal(err)
	}

	cfg := &format.Config{
		SearchDir: *searchDir,
		MainFile:  *generalInfo,
	}

	if err := format.New().Build(cfg); err != nil {
		log.Fatal(err)
	}
}
