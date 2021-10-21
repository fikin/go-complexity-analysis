package main

import (
	"github.com/fikin/go-complexity-analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(complexity.Analyzer)
}
