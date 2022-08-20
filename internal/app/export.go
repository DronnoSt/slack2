package app

import (
	"context"
	"errors"
	"fmt"
	"runtime/trace"
	"time"

	"github.com/rusq/slackdump/v2"
	"github.com/rusq/slackdump/v2/auth"
	"github.com/rusq/slackdump/v2/export"
	"github.com/rusq/slackdump/v2/fsadapter"
	"github.com/rusq/slackdump/v2/internal/mattermost"
	"github.com/rusq/slackdump/v2/logger"
)

// Export performs the full export of slack workspace in slack export compatible
// format.
func Export(ctx context.Context, cfg Config, prov auth.Provider) error {
	ctx, task := trace.NewTask(ctx, "Export")
	defer task.End()

	if cfg.ExportName == "" {
		return errors.New("export directory or filename not specified")
	}

	sess, err := slackdump.NewWithOptions(ctx, prov, cfg.Options)
	if err != nil {
		return err
	}

	expCfg := export.Options{
		Oldest:       time.Time(cfg.Oldest),
		Latest:       time.Time(cfg.Latest),
		IncludeFiles: cfg.Options.DumpFiles,
		Logger:       cfg.Logger(),
		List:         cfg.Input.List,
		Type:         cfg.ExportType,
	}
	switch cfg.ExportType {
	case export.TStandard:
		return runStandardExport(ctx, sess, cfg.ExportName, cfg.Logger(), expCfg)
	case export.TMattermost:
		return mattermost.Run(ctx, sess, cfg.TeamName, "", cfg.ExportName, expCfg)
	default:
		return fmt.Errorf("unknown export type: %s", cfg.ExportType)
	}
}

func runStandardExport(ctx context.Context, sess *slackdump.Session, exportName string, l logger.Interface, opts export.Options) error {
	fs, err := fsadapter.ForFilename(exportName)
	if err != nil {
		l.Debugf("Export:  filesystem error: %s", err)
		return fmt.Errorf("failed to initialise the filesystem: %w", err)
	}
	defer func() {
		l.Debugf("Export:  closing file system")
		if err := fsadapter.Close(fs); err != nil {
			l.Printf("Export:  error closing filesystem")
		}
	}()

	l.Debugf("Export:  filesystem: %s", fs)
	l.Printf("Export:  staring export to: %s", fs)

	e := export.New(sess, fs, opts)
	if err := e.Run(ctx); err != nil {
		return err
	}
	return nil
}
