package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fikin/go-complexity-analysis"
)

func TestIt(t *testing.T) {
	theConfig = &ConfigFile{}
	// outputFormat = "stylecheck"
	// assert.NoError(t, configureConfigIfGiven())
	// configureOutputFormat()
	funcsCnt := 0
	oldFnc := complexity.FuncStatsCallback
	complexity.FuncStatsCallback = func(s complexity.FuncStatsType) {
		funcsCnt++
		oldFnc(s)
	}
	assert.Equal(t, 1, run([]string{"./../../testdata/src/..."}, complexity.Analyzer))
	assert.Equal(t, 19, funcsCnt)
}
