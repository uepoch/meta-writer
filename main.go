package main

import (
	"github.com/uepoch/metrics-meta/cmd"
	"go.uber.org/zap"
)

func main() {
	l, _ := zap.NewProduction()
	zap.ReplaceGlobals(l)
	cmd.Execute()
}
