package a

import (
	"fmt"
	"log/slog"
	"net/http"

	"go.uber.org/zap"
)

func goodCases(logger *zap.Logger) {
	slog.Info("hello world")
	logger.Debug("hello world")
	logger.Warn("message with spaces")
	slog.Error(" there was an error")
}

func sensitiveCases(password string, apiToken string, logger *zap.Logger) {
	slog.Info("password: " + password)    // want "log message should not contain sensitive data"
	logger.Info("api_token: " + apiToken) // want "log message should not contain sensitive data"
}

func messageRules(logger *zap.Logger) {
	slog.Info("Hello world")    // want "first element should be a lower case char: H"
	logger.Error("error🔥")      // want "string should be contains only english chars"
	logger.Warn("   ")          // want "message should be not empty"
	slog.Info("Привет, мир")    // want "string should be contains only english chars"
	logger.Info("Привет, мир")  // want "string should be contains only english chars"
	slog.Error("postgres.New")  // want "string should be contains only english chars"
	logger.Warn("postgres.New") // want "string should be contains only english chars"
}

func nonLiteralMessage(logger *zap.Logger, msg string) {
	logger.Info(msg)
}

func targetDetectionBranches(logger *zap.Logger) {
	localInfo("hello world")
	slog.Default().Log(nil, slog.LevelInfo, "hello world")
	http.Error(nil, "", 200)
	(*zap.Logger).Info(logger, "hello world")
}

func containsSensitiveDataBranches(logger *zap.Logger) {
	logger.Info(fmt.Sprintf("code %d", 1))
}

func localInfo(msg string) {}
