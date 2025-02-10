package main

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func Test_ExitInMainAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, NoExitInMainAnalyzer,
		"../testdata/exit_in_goroutine",
		"../testdata/exit_in_defer",
		"../testdata/exit_in_select",
		"../testdata/main_exit",
		"../testdata/main_no_exit",
		"../testdata/not_main_func",
	)
}
