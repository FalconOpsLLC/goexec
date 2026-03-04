package smb

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/FalconOpsLLC/goexec/pkg/goexec"
	"github.com/rs/zerolog"
)

type FileStager struct {
	goexec.Cleaner

	Client *Client

	Share          string
	SharePath      string
	File           string
	relativePath   string
	ForceReconnect bool
}

func (o *FileStager) Upload(ctx context.Context, reader io.Reader) (err error) {

	// Calculate relative path from share root to target file (matches OutputFileFetcher pattern)
	shp := pathPrefix.ReplaceAllString(strings.ToLower(strings.ReplaceAll(o.SharePath, `\`, "/")), "")
	fp := pathPrefix.ReplaceAllString(strings.ToLower(strings.ReplaceAll(o.File, `\`, "/")), "")

	if o.relativePath, err = filepath.Rel(shp, fp); err != nil {
		return fmt.Errorf("calculate relative path: %w", err)
	}

	if o.ForceReconnect || !o.Client.connected {
		err = o.Client.Connect(ctx)
		if err != nil {
			return
		}
		defer o.AddCleaners(o.Client.Close)
	}

	if o.ForceReconnect || o.Client.share != o.Share {
		err = o.Client.Mount(ctx, o.Share)
		if err != nil {
			return
		}
	}

	writer, err := o.Client.mount.OpenFile(o.relativePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("open remote file for writing: %w", err)
	}

	if _, err = io.Copy(writer, reader); err != nil {
		return
	}

	o.AddCleaners(func(_ context.Context) error { return writer.Close() })

	return
}

// RemoveUploadedFile deletes the uploaded file from the remote filesystem.
// The share must already be mounted from a prior Upload call.
func (o *FileStager) RemoveUploadedFile(ctx context.Context) error {
	log := zerolog.Ctx(ctx)

	if o.Client.mount == nil {
		return fmt.Errorf("share not mounted")
	}

	if err := o.Client.mount.Remove(o.relativePath); err != nil {
		return fmt.Errorf("remove remote file: %w", err)
	}

	log.Info().Str("path", o.File).Msg("Removed uploaded file")
	return nil
}

// ConfirmUpload checks that the uploaded file exists on the remote filesystem
// and logs the file path and size. The share must already be mounted from a
// prior Upload call.
func (o *FileStager) ConfirmUpload(ctx context.Context) error {
	log := zerolog.Ctx(ctx)

	if o.Client.mount == nil {
		return fmt.Errorf("share not mounted")
	}

	info, err := o.Client.mount.Stat(o.relativePath)
	if err != nil {
		return fmt.Errorf("stat remote file: %w", err)
	}

	log.Info().
		Str("path", o.File).
		Int64("size", info.Size()).
		Msg("Upload confirmed")

	return nil
}
