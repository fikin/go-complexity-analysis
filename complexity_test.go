package complexity

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer is a test for Analyzer.
func TestAnalyzer(t *testing.T) {
	// asCsv = true
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer, []string{"a", "halstead"}...)
}
