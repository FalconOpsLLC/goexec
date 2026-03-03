package cmd

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "os"
  "time"

  "github.com/spf13/cobra"
  "github.com/spf13/pflag"
)

func registerLoggingFlags(fs *pflag.FlagSet) {
  fs.SortFlags = false
  fs.BoolVarP(&logDebug, "debug", "D", false, "Enable debug logging")
  fs.StringVarP(&logOutput, "log-file", "O", "", "Write JSON logging output to `file`")
  fs.BoolVarP(&logJson, "json", "j", false, "Write logging output in JSON lines")
  fs.BoolVarP(&logQuiet, "quiet", "q", false, "Disable info logging")
}

func registerNetworkFlags(fs *pflag.FlagSet) {
  fs.StringVarP(&proxy, "proxy", "x", "", "Proxy `URI`")
  fs.StringVarP(&rpcClient.Filter, "epm-filter", "F", "", "String `binding` to filter endpoints returned by the RPC endpoint mapper (EPM)")
  fs.StringVar(&rpcClient.Endpoint, "endpoint", "", "Explicit RPC endpoint string `binding`")
  fs.BoolVar(&rpcClient.NoEpm, "no-epm", false, "Don't use EPM to discover RPC endpoints")
  fs.BoolVar(&rpcClient.UseEpm, "epm", false, "Use EPM to discover available bindings")
  fs.BoolVar(&rpcClient.NoSign, "no-sign", false, "Disable signing on DCERPC messages")
  fs.BoolVar(&rpcClient.NoSeal, "no-seal", false, "Disable packet stub encryption on DCERPC messages")

  if err := fs.MarkHidden("no-epm"); err != nil {
    panic(err)
  }
  if err := fs.MarkDeprecated("no-epm", "use --epm=false instead"); err != nil {
    panic(err)
  }

  //cmd.MarkFlagsMutuallyExclusive("endpoint", "epm-filter")
  //cmd.MarkFlagsMutuallyExclusive("no-epm", "epm-filter")
}

func registerExecutionFlags(fs *pflag.FlagSet) {
  fs.StringVarP(&exec.Input.Executable, "exec", "e", "", "Remote Windows `executable` to invoke")
  fs.StringVarP(&exec.Input.Arguments, "args", "a", "", "Process command line arguments")
  fs.StringVarP(&exec.Input.Command, "command", "c", "", "Windows process command line (executable & arguments)")

  //cmd.MarkFlagsOneRequired("executable", "command")
  //cmd.MarkFlagsMutuallyExclusive("executable", "command")
}

func registerExecutionUploadFlags(fs *pflag.FlagSet) {
	fs.StringVar(&uploadSource, "upload", "", "Upload local `file` to remote filesystem")
	fs.StringVar(&uploadDest, "upload-dest", "", "Remote destination `path` for uploaded file")
	fs.BoolVar(&exec.Upload.NoDelete, "no-delete-upload", false, "Preserve uploaded file on remote filesystem")
}

func registerExecutionOutputFlags(fs *pflag.FlagSet) {
  fs.StringVarP(&outputPath, "out", "o", "", "Fetch execution output to `file` or \"-\" for standard output")
  fs.StringVarP(&outputMethod, "out-method", "m", "smb", "Method to fetch execution output")
  fs.DurationVar(&exec.Output.Timeout, "out-timeout", time.Second*60, "Output timeout `duration`")
  //fs.StringVar(&exec.Output.RemotePath, "out-remote", "", "Location to temporarily store output on remote filesystem")
  fs.BoolVar(&exec.Output.NoDelete, "no-delete-out", false, "Preserve output file on remote filesystem")
}

func args(reqs ...func(*cobra.Command, []string) error) (fn func(*cobra.Command, []string) error) {
  return func(cmd *cobra.Command, args []string) (err error) {

    for _, req := range reqs {
      if err = req(cmd, args); err != nil {
        return
      }
    }
    return
  }
}

func argsAcceptValues(name string, in *string, valid ...string) func(*cobra.Command, []string) error {
  return func(*cobra.Command, []string) error {
    for _, v := range valid {
      if *in == v {
        return nil
      }
    }
    if j, err := json.Marshal(valid); err == nil {
      return fmt.Errorf("parse %s: %q doesn't match any accepted values: %s", name, *in, string(j))
    } else {
      return err
    }
  }
}

func argsTarget(proto string) func(cmd *cobra.Command, args []string) error {

  return func(cmd *cobra.Command, args []string) (err error) {

    if len(args) != 1 {
      return errors.New("command require exactly one positional argument: [target]")
    }

    if credential, target, err = adAuthOpts.WithTarget(context.TODO(), proto, args[0]); err != nil {
      return fmt.Errorf("failed to parse target: %w", err)
    }

    if credential == nil {
      return errors.New("no credentials supplied")
    }
    if target == nil {
      return errors.New("no target supplied")
    }
    return
  }
}

func argsSmbClient() func(cmd *cobra.Command, args []string) error {
  return args(
    argsTarget("cifs"),

    func(_ *cobra.Command, _ []string) error {

      smbClient.Credential = credential
      smbClient.Target = target
      smbClient.Proxy = proxy

      return smbClient.Parse(context.TODO())
    },
  )
}

func argsRpcClient(proto string, endpoint string) func(cmd *cobra.Command, args []string) error {
  return args(
    argsTarget(proto),

    func(cmd *cobra.Command, args []string) (err error) {
      switch {
      case rpcClient.Endpoint != "":
      case rpcClient.Filter != "":
      case endpoint == "":
        rpcClient.UseEpm = true
      default:
        rpcClient.Endpoint = endpoint
      }
      rpcClient.Target = target
      rpcClient.Credential = credential
      rpcClient.Proxy = proxy

      return rpcClient.Parse(context.TODO())
    },
  )
}

func argsOutput(methods ...string) func(cmd *cobra.Command, args []string) error {

	var as []func(*cobra.Command, []string) error

	for _, method := range methods {
		if method == "smb" {
			as = append(as, argsSmbClient())
		}
	}

	return args(append(as, func(*cobra.Command, []string) (err error) {

		if outputPath != "" {
			if outputPath == "-" {
				exec.Output.Writer = os.Stdout

			} else if exec.Output.Writer, err = os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
				log.Fatal().Err(err).Msg("Failed to open output file")
			}
		}
		return
	})...)
}

func argsUpload(methods ...string) func(cmd *cobra.Command, args []string) error {

	var as []func(*cobra.Command, []string) error

	for _, method := range methods {
		if method == "smb" {
			as = append(as, argsSmbClient())
		}
	}

	return args(append(as, func(*cobra.Command, []string) (err error) {

		if uploadSource != "" {
			if uploadDest == "" {
				return fmt.Errorf("--upload-dest is required when --upload is set")
			}
			exec.Upload.Reader, err = os.Open(uploadSource)
			if err != nil {
				return fmt.Errorf("open upload file: %w", err)
			}
		}
		return
	})...)
}
