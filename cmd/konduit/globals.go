package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"runtime"
)

type Globals struct {
	Version VersionCmd `cmd:"" help:"Print version information."`

	Log    Log       `embed:"" prefix:"log." envprefix:"LOG_"`
	Stdout io.Writer `kong:"-"`
	Stderr io.Writer `kong:"-"`
}

type Log struct {
	*slog.Logger

	Level  string `env:"LEVEL" default:"info" enum:"debug,info,warn,error" help:"Configure the log level."`
	Format string `env:"FORMAT" default:"text" enum:"text,json" help:"Configure the log format."`
}

func (l *Log) AfterApply(g *Globals) error {
	opts := new(slog.HandlerOptions)

	switch l.Level {
	case "debug":
		opts.Level = slog.LevelDebug
	case "info":
		opts.Level = slog.LevelInfo
	case "warn":
		opts.Level = slog.LevelWarn
	case "error":
		opts.Level = slog.LevelError
	default:
		return fmt.Errorf("invalid log level: %s", l.Level)
	}

	var handler slog.Handler
	switch l.Format {
	case "text":
		handler = slog.NewTextHandler(g.Stderr, opts)
	case "json":
		handler = slog.NewJSONHandler(g.Stderr, opts)
	default:
		return fmt.Errorf("invalid log format: %s", l.Format)
	}

	l.Logger = slog.New(handler)
	return nil
}

var (
	version = "dev"
	commit  = "unknown"
)

type BuildInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	GoVersion string `json:"goVersion"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

type VersionCmd struct{}

func (c *VersionCmd) Run(ctx context.Context, g *Globals) error {
	info := BuildInfo{
		Version:   version,
		GitCommit: commit,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
	enc := json.NewEncoder(g.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(info); err != nil {
		return fmt.Errorf("encode build info: %w", err)
	}
	return nil
}
