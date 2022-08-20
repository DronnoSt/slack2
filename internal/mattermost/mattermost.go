// Package mattermost handles the generation of mattermost compatible export
// file (EXPERIMENTAL).
package mattermost

import (
	"archive/zip"
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rusq/slackdump/v2"
	"github.com/rusq/slackdump/v2/export"
	"github.com/rusq/slackdump/v2/fsadapter"
	"github.com/rusq/slackdump/v2/logger"

	mmslack "github.com/mattermost/mmetl/services/slack"
)

func Run(ctx context.Context, sd *slackdump.Session, teamName, tmpdir, zipFile string, opts export.Options) error {
	tmpdir, err := os.MkdirTemp(tmpdir, "slackdump*")
	if err != nil {
		return err
	}
	logger.Default.Print(tmpdir)

	zipfile, err := os.CreateTemp(tmpdir, "slackdump*.zip")
	if err != nil {
		return err
	}
	if err := zipfile.Close(); err != nil {
		return err
	}

	zfs, err := fsadapter.NewZipFile(zipfile.Name())
	if err != nil {
		return err
	}
	opts.Type = export.TMattermost // enforce mattermost export type

	exp := export.New(sd, zfs, opts)
	if err := exp.Run(ctx); err != nil {
		zfs.Close()
		return err
	}
	if err := zfs.Close(); err != nil {
		return err
	}

	fi, err := os.Stat(zipfile.Name())
	if err != nil {
		return err
	}
	expF, err := os.Open(zipfile.Name())
	if err != nil {
		return err
	}
	defer expF.Close()

	zr, err := zip.NewReader(expF, fi.Size())
	if err != nil {
		return err
	}

	st, err := mmslack.ParseSlackExportFile(teamName, zr, false)
	if err != nil {
		return err
	}

	mmdir := filepath.Join(tmpdir, "mm")
	mmfiledir := filepath.Join(mmdir, "bulk-export-attachments")
	mmexpfile := filepath.Join(mmdir, "mattermost_import.jsonl")

	if err := os.MkdirAll(mmfiledir, 0700); err != nil {
		return err
	}

	im, err := mmslack.Transform(st, mmfiledir, false)
	if err != nil {
		return err
	}
	if err := mmslack.Export(teamName, im, mmexpfile); err != nil {
		return err
	}

	out, err := os.Create(zipFile)
	if err != nil {
		return err
	}
	defer out.Close()

	zw := zip.NewWriter(out)
	defer zw.Close()
	if err := filepath.WalkDir(mmdir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(mmdir, path)
		if err != nil {
			return err
		}
		if filepath.Dir(relPath) == "bulk-export-attachments" {
			relPath = filepath.Join("data", relPath)
		}
		w, err := zw.Create(relPath)
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := io.Copy(w, f); err != nil {
			return err
		}
		logger.Default.Printf("packed: %s", relPath)
		return nil
	}); err != nil {
		return err
	}
	return nil
}
