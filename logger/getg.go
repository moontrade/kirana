//go:build (amd64 || arm64) && go1.20

package logger

func getg() *g
