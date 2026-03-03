package cmd

import (
  "context"
  "encoding/json"
  "github.com/FalconOpsLLC/goexec/pkg/goexec"
  wmiexec "github.com/FalconOpsLLC/goexec/pkg/goexec/wmi"
  "github.com/oiweiwei/go-msrpc/ssp/gssapi"
  "github.com/spf13/cobra"
  "os"
)

func wmiCmdInit() {
  cmdFlags[wmiCmd] = []*flagSet{
    defaultAuthFlags,
    defaultLogFlags,
    defaultNetRpcFlags,
  }
  wmiCallCmdInit()
  wmiProcCmdInit()

  wmiCmd.PersistentFlags().AddFlagSet(defaultAuthFlags.Flags)
  wmiCmd.PersistentFlags().AddFlagSet(defaultLogFlags.Flags)
  wmiCmd.PersistentFlags().AddFlagSet(defaultNetRpcFlags.Flags)
  wmiCmd.AddCommand(wmiProcCmd, wmiCallCmd)
}

func wmiCallCmdInit() {
  wmiCallFlags := newFlagSet("WMI")

  wmiCallFlags.Flags.StringVarP(&wmiCall.Resource, "namespace", "n", "//./root/cimv2", "WMI namespace")
  wmiCallFlags.Flags.StringVarP(&wmiCall.Class, "class", "C", "", `WMI class to instantiate (i.e. "Win32_Process")`)
  wmiCallFlags.Flags.StringVarP(&wmiCall.Method, "method", "m", "", `WMI Method to call (i.e. "Create")`)
  wmiCallFlags.Flags.StringVarP(&wmiArguments, "args", "A", "{}", `WMI Method argument(s) in JSON dictionary format (i.e. {"Command":"calc.exe"})`)

  wmiCallCmd.Flags().AddFlagSet(wmiCallFlags.Flags)

  cmdFlags[wmiCallCmd] = []*flagSet{
    wmiCallFlags,
    defaultAuthFlags,
    defaultLogFlags,
    defaultNetRpcFlags,
  }
  if err := wmiCallCmd.MarkFlagRequired("class"); err != nil {
    panic(err)
  }
  if err := wmiCallCmd.MarkFlagRequired("method"); err != nil {
    panic(err)
  }
}

func wmiProcCmdInit() {
	wmiProcExecFlags := newFlagSet("Execution")

	registerExecutionFlags(wmiProcExecFlags.Flags)
	registerExecutionOutputFlags(wmiProcExecFlags.Flags)
	registerExecutionUploadFlags(wmiProcExecFlags.Flags)

	wmiProcExecFlags.Flags.StringVarP(&wmiProc.WorkingDirectory, "directory", "d", `C:\`, "Working directory")

  cmdFlags[wmiProcCmd] = []*flagSet{
    wmiProcExecFlags,
    defaultAuthFlags,
    defaultLogFlags,
    defaultNetRpcFlags,
  }

  wmiProcCmd.Flags().AddFlagSet(wmiProcExecFlags.Flags)
}

var (
  wmiCall = wmiexec.WmiCall{}
  wmiProc = wmiexec.WmiProc{}

  wmiArguments string

  wmiCmd = &cobra.Command{
    Use:   "wmi",
    Short: "Execute with Windows Management Instrumentation (MS-WMI)",
    Long: `Description:
  The wmi module uses remote Windows Management Instrumentation (WMI) to
  perform various operations including process creation.`,
    GroupID: "module",
    Args:    cobra.NoArgs,
  }

  wmiCallCmd = &cobra.Command{
    Use:   "call [target]",
    Short: "Execute specified WMI method",
    Long: `Description:
  The call method creates an instance of the specified WMI class (-c),
  then calls the provided method (-m) with the provided arguments (-A).`,
    Args: args(
      argsRpcClient("cifs", ""),
      func(cmd *cobra.Command, args []string) error {
        return json.Unmarshal([]byte(wmiArguments), &wmiCall.Args)
      }),

    Run: func(cmd *cobra.Command, args []string) {
      wmiCall.Client = &rpcClient
      wmiCall.Out = os.Stdout

      ctx := log.With().
        Str("module", "wmi").
        Str("method", "call").
        Logger().WithContext(gssapi.NewSecurityContext(context.Background()))

      if err := goexec.ExecuteCleanAuxiliaryMethod(ctx, &wmiCall); err != nil {
        log.Fatal().Err(err).Msg("Operation failed")
      }
    },
  }

  wmiProcCmd = &cobra.Command{
    Use:   "proc [target]",
    Short: "Start a Windows process",
    Long: `Description:
  The proc method creates an instance of the Win32_Process WMI class, then
  calls the Win32_Process.Create method with the provided command (-c),
  and optional working directory (-d).`,
		Args: args(
			argsRpcClient("cifs", ""),
			argsOutput("smb"),
			argsUpload("smb"),
		),

		Run: func(cmd *cobra.Command, args []string) {
			wmiProc.Client = &rpcClient
      wmiProc.IO = exec
      wmiProc.Resource = "//./root/cimv2"

      ctx := log.With().
        Str("module", "wmi").
        Str("method", "proc").
        Logger().WithContext(gssapi.NewSecurityContext(context.Background()))

      if err := goexec.ExecuteCleanMethod(ctx, &wmiProc, &exec); err != nil {
        log.Fatal().Err(err).Msg("Operation failed")
      }
    },
  }
)
