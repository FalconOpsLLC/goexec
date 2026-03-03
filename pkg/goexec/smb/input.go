package smb

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/FalconOpsLLC/goexec/pkg/goexec"
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
