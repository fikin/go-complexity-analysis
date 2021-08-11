package main

import (
	"github.com/fikin/go-complexity-analysis"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() { unitchecker.Main(complexity.Analyzer) }
