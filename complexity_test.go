package complexity

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer is a test for Analyzer.
func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer, []string{"a", "halstead"}...)
}

// TestAnalyzer2 is a test for Analyzer.
func TestAnalyzer2(t *testing.T) {
	testdata := analysistest.TestData()
	// asCsv = true
	analysistest.Run(t, testdata, Analyzer, []string{"b"}...)
}
