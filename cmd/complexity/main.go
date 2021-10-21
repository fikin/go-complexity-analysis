package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fikin/go-complexity-analysis"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/packages"
)

func main() {
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

	configureOutputFormat()

	os.Exit(run(args, a))
}

type checkstyleErrorTag struct {
	XMLName  xml.Name `xml:"error"`
	Col      int      `xml:"column,attr"`
	Line     int      `xml:"line,attr"`
	Msg      string   `xml:"message,attr"`
	Severity string   `xml:"severity,attr,omitempty"`
	Source   string   `xml:"source,attr,omitempty"`
}
type checkstyleFileTag struct {
	XMLName  xml.Name `xml:"file"`
	FileName string   `xml:"name,attr"`
	Errors   []checkstyleErrorTag
}

// checkstyleTag is structure used to serialize in xml all diagnostic
type checkstyleTag struct {
	XMLName    xml.Name `xml:"checkstyle"`
	Version    string   `xml:"version,attr"`
	filesAsMap map[string]checkstyleFileTag
	Files      []checkstyleFileTag
}

// flag option when in standalone cmdline mode
// one of : txt, csv, checkstyle
var outputFormat = "txt"

// gathered function stats to be printed at the end when output-format=csv
var funcStats = []complexity.FuncStatsType{}

// gathered function stats to be printed at the end when output-format=stylechek
var checkstyles = checkstyleTag{filesAsMap: map[string]checkstyleFileTag{}, Files: []checkstyleFileTag{}, Version: "5.0"}

func addCmdlineFlags(a *analysis.Analyzer) {
	flag.StringVar(&outputFormat, "output-format", "txt", "to print the diagnostics as 'csv', 'checkstyle' xml or vet-like 'txt' (default 'txt')")
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
					i = checkstyleFileTag{FileName: stats.Filename, Errors: []checkstyleErrorTag{}}
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

func run(args []string, analyzer *analysis.Analyzer) (exitcode int) {
	pkg, err := load(args)
	if err != nil {
		log.Print(err)
		return 1 // load errors
	}

	analyzers := deepScanRequires(analyzer)

	foundDiagnostics := analyze(pkg, analyzers)

	printDiagnostics(foundDiagnostics)

	if len(foundDiagnostics) > 0 {
		return 1
	}
	return 0

}

// deepScanRequires deep-scans Requires fields and returns the ordered array of analyzers
func deepScanRequires(analyzer *analysis.Analyzer) []*analysis.Analyzer {
	if analyzer == nil {
		return nil
	}
	// TODO : strictly speaking append must de-duplicate analyzers i.e. use set instead of array
	arr := []*analysis.Analyzer{}
	for _, d := range analyzer.Requires {
		arr = append(arr, deepScanRequires(d)...)
	}
	return append(arr, analyzer)
}

// load loads the initial packages.
func load(patterns []string) ([]*packages.Package, error) {
	conf := packages.Config{
		Mode:  packages.LoadSyntax,
		Tests: true,
	}
	initial, err := packages.Load(&conf, patterns...)
	if err == nil {
		if n := packages.PrintErrors(initial); n > 1 {
			err = fmt.Errorf("%d errors during loading", n)
		} else if n == 1 {
			err = fmt.Errorf("error during loading")
		} else if len(initial) == 0 {
			err = fmt.Errorf("%s matched no packages", strings.Join(patterns, " "))
		}
	}

	return initial, err
}

type analyzerResultsType map[*analysis.Analyzer]interface{}

type foundDiagnosticsStruct struct {
	pkg         *packages.Package
	diagnostics []analysis.Diagnostic
	err         error
}

func analyze(pkgs []*packages.Package, analyzers []*analysis.Analyzer) []foundDiagnosticsStruct {
	d := []foundDiagnosticsStruct{}
	for _, pkg := range pkgs {
		analyzerResults := analyzerResultsType{}
		for _, a := range analyzers {
			diags, err := analyzePkg(&analyzerResults, pkg, a)
			d = append(d, foundDiagnosticsStruct{pkg: pkg, diagnostics: diags, err: err})
		}
	}
	return d
}

func analyzePkg(results *analyzerResultsType, pkg *packages.Package, a *analysis.Analyzer) ([]analysis.Diagnostic, error) {
	diagnostics := []analysis.Diagnostic{}
	pass := &analysis.Pass{
		Analyzer:          a,
		Fset:              pkg.Fset,
		Files:             pkg.Syntax,
		OtherFiles:        pkg.OtherFiles,
		IgnoredFiles:      pkg.IgnoredFiles,
		Pkg:               pkg.Types,
		TypesInfo:         pkg.TypesInfo,
		TypesSizes:        pkg.TypesSizes,
		ResultOf:          *results,
		Report:            func(d analysis.Diagnostic) { diagnostics = append(diagnostics, d) },
		ImportObjectFact:  nil,
		ExportObjectFact:  nil,
		ImportPackageFact: nil,
		ExportPackageFact: nil,
		AllObjectFacts:    nil,
		AllPackageFacts:   nil,
	}
	res, err := a.Run(pass)
	if err == nil {
		(*results)[a] = res
	}
	return diagnostics, err
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

func doPrintDiagnostics(arr []foundDiagnosticsStruct) {
	for _, f := range arr {
		if f.err != nil {
			fmt.Printf("%s : %v\n", f.pkg.Name, f.err)
		}
		for _, d := range f.diagnostics {
			fmt.Printf("%s : %d : %s\n", f.pkg.Name, d.Pos, d.Message)
		}
	}
}

func doPrintFuncStats(arr []complexity.FuncStatsType) {
	for _, stats := range arr {
		if stats.IsNotMaintenable || stats.IsTooComplex {
			fmt.Printf("%s,%d,%s,%d,%d,%0.3f,%0.3f,%0.3f,%d,%d,%t,%t\n",
				stats.Filename, stats.Line, stats.FunctionName,
				stats.CyclomaticComplexity, stats.MaintenabilityIndex, stats.HalsbreadDifficulty,
				stats.HalsbreadVolume, stats.TimeToCode,
				stats.LOC, stats.ConstantsLOC,
				stats.IsTooComplex, stats.IsNotMaintenable)
		}
	}
}

func doPrintcheckstyles(data checkstyleTag) {
	for _, v := range data.filesAsMap {
		data.Files = append(data.Files, v)
	}
	output, err := xml.MarshalIndent(data, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	os.Stdout.Write(output)
}
