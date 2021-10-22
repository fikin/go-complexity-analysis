package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fikin/go-complexity-analysis"
	"golang.org/x/tools/go/analysis"
)

// flag option only in standalone cmdline mode
// one of : txt, csv, checkstyle
var outputFormat = "txt"

// flag option only standalone cmdline mode
// its format is golangci-lint like yaml configuration
// subject to limited flags support (see README)
var configfile string

// gathered function stats to be printed at the end when output-format=csv
var funcStats = []complexity.FuncStatsType{}

// gathered function stats to be printed at the end when output-format=stylechek
var checkstyles = checkstyleTag{filesAsMap: map[string]checkstyleFileTag{}, Files: []checkstyleFileTag{}, Version: "5.0"}

var currDir string

func main() {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	currDir = path

	a := complexity.Analyzer
	log.SetFlags(0)
	log.SetPrefix(a.Name + ": ")

	analyzers := []*analysis.Analyzer{a}

	if err := analysis.Validate(analyzers); err != nil {
		log.Fatal(err)
	}

	addCmdlineFlags(a)

	flag.Parse() // (ExitOnError)

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	configureConfigIfGiven()
	configureOutputFormat()

	os.Exit(run(args, a))
}

func addCmdlineFlags(a *analysis.Analyzer) {
	flag.StringVar(&outputFormat, "out-format", "txt", "to print the diagnostics as 'csv', 'checkstyle' xml or vet-like 'txt' (default 'txt')")
	flag.StringVar(&configfile, "c", "", "configuration like golangci")
	flag.Usage = func() {
		paras := strings.Split(a.Doc, "\n\n")
		fmt.Fprintf(os.Stderr, "%s: %s\n\n", a.Name, paras[0])
		fmt.Fprintf(os.Stderr, "Usage: %s [-flag] [package]\n\n", a.Name)
		if len(paras) > 1 {
			fmt.Fprintln(os.Stderr, strings.Join(paras[1:], "\n\n"))
		}
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flag.PrintDefaults()
	}
}

func configureOutputFormat() {
	switch outputFormat {
	case "checkstyle":
		complexity.FuncStatsCallback = func(stats complexity.FuncStatsType) {
			msg := complexity.ToDiagnosticMsg(stats)
			if msg != "" {
				i, ok := checkstyles.filesAsMap[stats.Filename]
				if !ok {
					i = checkstyleFileTag{FileName: getRelativeFileName(stats.Filename, currDir), Errors: []checkstyleErrorTag{}}
				}
				i.Errors = append(i.Errors, checkstyleErrorTag{Line: stats.Line, Msg: msg, Severity: "error", Source: "typecheck"})
				checkstyles.filesAsMap[stats.Filename] = i
			}
		}
	case "csv":
		complexity.FuncStatsCallback = func(stats complexity.FuncStatsType) {
			funcStats = append(funcStats, stats)
		}
	}
}

func printDiagnostics(arr []foundDiagnosticsStruct) {
	switch outputFormat {
	case "checkstyle":
		doPrintcheckstyles(checkstyles)
	case "csv":
		doPrintFuncStats(funcStats)
	default:
		doPrintDiagnostics(arr)
	}
}

func doPrintFuncStats(arr []complexity.FuncStatsType) {
	for _, stats := range arr {
		if stats.IsNotMaintenable || stats.IsTooComplex {
			fmt.Printf("%s,%d,%s,%d,%d,%0.3f,%0.3f,%0.3f,%d,%d,%t,%t\n",
				getRelativeFileName(stats.Filename, currDir), stats.Line, stats.FunctionName,
				stats.CyclomaticComplexity, stats.MaintenabilityIndex, stats.HalsbreadDifficulty,
				stats.HalsbreadVolume, stats.TimeToCode,
				stats.LOC, stats.ConstantsLOC,
				stats.IsTooComplex, stats.IsNotMaintenable)
		}
	}
}

func getRelativeFileName(filename string, basePath string) string {
	if strings.HasPrefix(filename, basePath+"/") {
		return filename[len(basePath)+1:]
	}
	return filename
}
