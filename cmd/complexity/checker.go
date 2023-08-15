package main

import (
	"fmt"
	"log"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/packages"
)

type analyzerResultsType map[*analysis.Analyzer]interface{}

type foundDiagnosticsStruct struct {
	pkg         *packages.Package
	diagnostics []analysis.Diagnostic
	err         error
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

// load loads the packages.
func load(patterns []string) ([]*packages.Package, error) {
	conf := packages.Config{
		// nolint:staticcheck
		Mode:       packages.LoadSyntax,
		Tests:      theConfig.Run.Tests,
		BuildFlags: formBuildTags(theConfig.Run.BuildTags),
	}
	pkgs, err := packages.Load(&conf, patterns...)
	if err == nil {
		if n := packages.PrintErrors(pkgs); n > 1 {
			err = fmt.Errorf("%d errors during loading", n)
		} else if n == 1 {
			err = fmt.Errorf("error during loading")
		} else if len(pkgs) == 0 {
			err = fmt.Errorf("%s matched no packages", strings.Join(patterns, " "))
		}
	}

	return pkgs, err
}

func formBuildTags(buildTags []string) []string {
	if len(buildTags) == 0 {
		return buildTags
	}
	return []string{"--tags", strings.Join(buildTags, ",")}
}

func analyze(pkgs []*packages.Package, analyzers []*analysis.Analyzer) []foundDiagnosticsStruct {
	d := []foundDiagnosticsStruct{}
	for _, pkg := range pkgs {
		analyzerResults := analyzerResultsType{}
		for _, a := range analyzers {
			diags, err := analyzePkg(&analyzerResults, pkg, a)
			if err != nil || len(diags) > 0 {
				d = append(d, foundDiagnosticsStruct{pkg: pkg, diagnostics: diags, err: err})
			}
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
