package main

import (
	"go.uber.org/zap"
)

func main() {
	sugar := zap.NewExample().Sugar()
	sugar.Info("Hello")
}
