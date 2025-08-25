package log

import (
	"github.com/dresswithpockets/openstats/app/env"
	"github.com/rotisserie/eris"
	"log/slog"
	"os"
)

var Logger *slog.Logger

var SlogLevelMap = map[string]slog.Level{
	"Debug": slog.LevelDebug,
	"Info":  slog.LevelInfo,
	"Warn":  slog.LevelWarn,
	"Error": slog.LevelError,
}

func Setup() error {
	logLevel, matchedErr := env.GetMapped("OPENSTATS_SLOG_LEVEL", SlogLevelMap)
	if matchedErr != nil {
		return matchedErr
	}

	handlerOptions := &slog.HandlerOptions{Level: logLevel}

	slogMode := env.GetString("OPENSTATS_SLOG_MODE")
	switch slogMode {
	case "Text":
		Logger = slog.New(slog.NewTextHandler(os.Stdout, handlerOptions))
	case "JSON":
		Logger = slog.New(slog.NewJSONHandler(os.Stdout, handlerOptions))
	default:
		return eris.Errorf("invalid value for OPENSTATS_SLOG_MODE: %s", slogMode)
	}

	return nil
}
