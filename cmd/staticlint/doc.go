// Package main реализует multichecker для статического анализа кода.
//
// Multichecker включает:
// - Стандартные анализаторы из golang.org/x/tools/go/analysis/passes.
// - Анализаторы SA из staticcheck.io.
// - Дополнительный анализатор staticcheck ST1000.
// - Публичные анализаторы errcheck и bodyclose.
// - Собственный анализатор noExitInMain, запрещающий вызов os.Exit в main.
//
// Запуск:
//
//	go run cmd/staticlint/main.go ./...
package main
