package goexec

import (
  "context"
  "fmt"
  "io"
  "strings"
  "time"
)

type OutputProvider interface {
	GetOutput(ctx context.Context, writer io.Writer) (err error)
	Clean(ctx context.Context) (err error)
}

type InputProvider interface {
	Upload(ctx context.Context, reader io.Reader) (err error)
	Clean(ctx context.Context) (err error)
}

type ExecutionIO struct {
	Cleaner

	Input  *ExecutionInput
	Output *ExecutionOutput
	Upload *ExecutionUpload
}

type ExecutionOutput struct {
	NoDelete   bool
	RemotePath string
	Timeout    time.Duration
	Provider   OutputProvider
	Writer     io.WriteCloser
}

type ExecutionUpload struct {
	NoConfirm  bool
	RemotePath string
	Provider   InputProvider
	Reader     io.ReadCloser
}

// UploadConfirmer is an optional interface that InputProvider implementations
// can satisfy to confirm a file was successfully uploaded.
type UploadConfirmer interface {
	ConfirmUpload(ctx context.Context) error
}

type ExecutionInput struct {
  StageFile      io.ReadCloser
  Executable     string
  ExecutablePath string
  Arguments      string
  Command        string
}

func (execIO *ExecutionIO) DoUpload(ctx context.Context) (err error) {
	if execIO.Upload != nil && execIO.Upload.Provider != nil && execIO.Upload.Reader != nil {
		return execIO.Upload.Provider.Upload(ctx, execIO.Upload.Reader)
	}
	return nil
}

func (execIO *ExecutionIO) CleanUpload(ctx context.Context) (err error) {
	if execIO.Upload != nil && execIO.Upload.Provider != nil {
		return execIO.Upload.Provider.Clean(ctx)
	}
	return nil
}

func (execIO *ExecutionIO) GetOutput(ctx context.Context) (err error) {
  if execIO.Output.Provider != nil {
    ctx = context.WithValue(ctx, ContextOptionOutputTimeout, execIO.Output.Timeout)
    return execIO.Output.Provider.GetOutput(ctx, execIO.Output.Writer)
  }
  return nil
}

func (execIO *ExecutionIO) Clean(ctx context.Context) (err error) {
  if execIO.Output.Provider != nil {
    return execIO.Output.Provider.Clean(ctx)
  }
  return nil
}

func (execIO *ExecutionIO) CommandLine() (cmd []string) {
  if execIO.Output.Provider != nil && execIO.Output.RemotePath != "" {
    return []string{
      `C:\Windows\System32\cmd.exe`,
      fmt.Sprintf(`/C%s >%s 2>&1`, execIO.Input.String(), execIO.Output.RemotePath),
    }
  }
  return execIO.Input.CommandLine()
}

func (execIO *ExecutionIO) String() (str string) {
  cmd := execIO.CommandLine()
  // Ensure that executable paths are quoted
  if strings.Contains(cmd[0], " ") {
    str = fmt.Sprintf(`%q %s`, cmd[0], strings.Join(cmd[1:], " "))
  } else {
    str = strings.Join(cmd, " ")
  }
  return strings.TrimSpace(str) // trim whitespace
}

func (i *ExecutionInput) CommandLine() (cmd []string) {
  cmd = make([]string, 2)
  cmd[1] = i.Arguments

  switch {
  case i.Command != "":
    copy(cmd, strings.SplitN(i.Command, " ", 2))
  case i.ExecutablePath != "":
    cmd[0] = i.ExecutablePath
  case i.Executable != "":
    cmd[0] = i.Executable
  }

  return cmd
}

func (i *ExecutionInput) String() string {
  return strings.TrimSpace(strings.Join(i.CommandLine(), " "))
}

func (i *ExecutionInput) Reader() (reader io.Reader) {
  if i.StageFile != nil {
    return i.StageFile
  }
  return strings.NewReader(i.String())
}
