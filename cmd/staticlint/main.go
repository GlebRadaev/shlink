package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"

	"github.com/kisielk/errcheck/errcheck"
	"github.com/timakin/bodyclose/passes/bodyclose"
)

func main() {
	var analyzers []*analysis.Analyzer

	// Добавляем стандартные анализаторы
	analyzers = append(analyzers,
		shadow.Analyzer,
		fieldalignment.Analyzer,
		nilfunc.Analyzer,
	)

	// Добавляем все SA анализаторы из staticcheck
	for _, v := range staticcheck.Analyzers {
		if v.Analyzer.Name[:2] == "SA" {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	for _, v := range simple.Analyzers {
		if v.Analyzer.Name == "S1030" {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	// Добавляем публичные анализаторы
	analyzers = append(analyzers, errcheck.Analyzer)
	analyzers = append(analyzers, bodyclose.Analyzer)

	// Добавляем собственный анализатор
	analyzers = append(analyzers, NoExitInMainAnalyzer)

	// Запускаем multichecker
	multichecker.Main(analyzers...)
}
