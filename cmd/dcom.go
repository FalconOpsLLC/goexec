package cmd

import (
  "context"
  "fmt"
  "io"
  "os"
  "strings"

  "github.com/FalconOpsLLC/goexec/pkg/goexec"
  dcomexec "github.com/FalconOpsLLC/goexec/pkg/goexec/dcom"
  "github.com/oiweiwei/go-msrpc/ssp/gssapi"
  "github.com/spf13/cobra"
)

func dcomCmdInit() {
  cmdFlags[dcomCmd] = []*flagSet{
    defaultAuthFlags,
    defaultLogFlags,
    defaultNetRpcFlags,
  }
  dcomMmcCmdInit()
  dcomShellWindowsCmdInit()
  dcomShellBrowserWindowCmdInit()
  dcomHtafileCmdInit()
  dcomExcelCmdInit()
  dcomVisualStudioCmdInit()

  dcomCmd.PersistentFlags().AddFlagSet(defaultAuthFlags.Flags)
  dcomCmd.PersistentFlags().AddFlagSet(defaultLogFlags.Flags)
  dcomCmd.PersistentFlags().AddFlagSet(defaultNetRpcFlags.Flags)
  dcomCmd.AddCommand(
    dcomMmcCmd,
    dcomShellWindowsCmd,
    dcomShellBrowserWindowCmd,
    dcomHtafileCmd,
    dcomExcelCmd,
    dcomVisualStudioCmd,
  )
}

func dcomExcelCmdInit() {
  cmdFlags[dcomExcelCmd] = []*flagSet{
    defaultAuthFlags,
    defaultLogFlags,
    defaultNetRpcFlags,
  }
  dcomExcelMacroCmdInit()
  dcomExcelXllCmdInit()
  dcomExcelCmd.AddCommand(dcomExcelMacroCmd, dcomExcelXllCmd)
}

func dcomVisualStudioCmdInit() {
  cmdFlags[dcomVisualStudioCmd] = []*flagSet{
    defaultAuthFlags,
    defaultLogFlags,
    defaultNetRpcFlags,
  }
  dcomVisualStudioDteCmdInit()
  dcomVisualStudioCmd.AddCommand(dcomVisualStudioDteCmd)
}

func dcomMmcCmdInit() {
	dcomMmcExecFlags := newFlagSet("Execution")
	registerExecutionFlags(dcomMmcExecFlags.Flags)
	registerExecutionOutputFlags(dcomMmcExecFlags.Flags)
	registerExecutionUploadFlags(dcomMmcExecFlags.Flags)
	dcomMmcExecFlags.Flags.StringVar(&dcomMmc.WorkingDirectory, "directory", `C:\`, "Working `directory`")
	dcomMmcExecFlags.Flags.StringVar(&dcomMmc.WindowState, "window", "Minimized", "Window state")

	cmdFlags[dcomMmcCmd] = []*flagSet{
		dcomMmcExecFlags,
		defaultAuthFlags,
		defaultLogFlags,
		defaultNetRpcFlags,
	}
	dcomMmcCmd.Flags().AddFlagSet(dcomMmcExecFlags.Flags)

	// Constraints
	dcomMmcCmd.MarkFlagsOneRequired("command", "exec", "upload")
}

func dcomShellWindowsCmdInit() {
	dcomShellWindowsExecFlags := newFlagSet("Execution")
	registerExecutionFlags(dcomShellWindowsExecFlags.Flags)
	registerExecutionOutputFlags(dcomShellWindowsExecFlags.Flags)
	registerExecutionUploadFlags(dcomShellWindowsExecFlags.Flags)
	dcomShellWindowsExecFlags.Flags.StringVar(&dcomShellWindows.WorkingDirectory, "directory", `C:\`, "Working directory `path`")
	dcomShellWindowsExecFlags.Flags.StringVar(&dcomShellWindows.WindowState, "app-window", "0", "Application window state `ID`")

	cmdFlags[dcomShellWindowsCmd] = []*flagSet{
		dcomShellWindowsExecFlags,
		defaultAuthFlags,
		defaultLogFlags,
		defaultNetRpcFlags,
	}
	dcomShellWindowsCmd.Flags().AddFlagSet(dcomShellWindowsExecFlags.Flags)

	// Constraints
	dcomShellWindowsCmd.MarkFlagsOneRequired("command", "exec", "upload")
}

func dcomShellBrowserWindowCmdInit() {
	dcomShellBrowserWindowExecFlags := newFlagSet("Execution")
	registerExecutionFlags(dcomShellBrowserWindowExecFlags.Flags)
	registerExecutionOutputFlags(dcomShellBrowserWindowExecFlags.Flags)
	registerExecutionUploadFlags(dcomShellBrowserWindowExecFlags.Flags)
	dcomShellBrowserWindowExecFlags.Flags.StringVar(&dcomShellBrowserWindow.WorkingDirectory, "directory", `C:\`, "Working directory `path`")
	dcomShellBrowserWindowExecFlags.Flags.StringVar(&dcomShellBrowserWindow.WindowState, "app-window", "0", "Application window state `ID`")

	cmdFlags[dcomShellBrowserWindowCmd] = []*flagSet{
		dcomShellBrowserWindowExecFlags,
		defaultAuthFlags,
		defaultLogFlags,
		defaultNetRpcFlags,
	}
	dcomShellBrowserWindowCmd.Flags().AddFlagSet(dcomShellBrowserWindowExecFlags.Flags)

	// Constraints
	dcomShellBrowserWindowCmd.MarkFlagsOneRequired("command", "exec", "upload")
}

func dcomHtafileCmdInit() {
	dcomHtafileExecFlags := newFlagSet("Execution")
	dcomHtafileExecFlags.Flags.StringVarP(&dcomHtafile.Url, "url", "U", "", "Load custom `URL`")
	dcomHtafileExecFlags.Flags.StringVar(&dcomHtafile.Javascript, "js", "", "Execute JavaScript one-liner")
	dcomHtafileExecFlags.Flags.StringVar(&dcomHtafile.Vbscript, "vbs", "", "Execute VBScript one-liner")
	registerExecutionFlags(dcomHtafileExecFlags.Flags)
	registerExecutionOutputFlags(dcomHtafileExecFlags.Flags)
	registerExecutionUploadFlags(dcomHtafileExecFlags.Flags)

	cmdFlags[dcomHtafileCmd] = []*flagSet{
		dcomHtafileExecFlags,
		defaultAuthFlags,
		defaultLogFlags,
		defaultNetRpcFlags,
	}
	dcomHtafileCmd.Flags().AddFlagSet(dcomHtafileExecFlags.Flags)

	// Constraints
	dcomHtafileCmd.MarkFlagsOneRequired("command", "exec", "url", "js", "vbs", "upload")
}

func dcomExcelMacroCmdInit() {
	dcomExcelMacroExecFlags := newFlagSet("Execution")
	dcomExcelMacroExecFlags.Flags.StringArrayVarP(&dcomExcelMacro.Macros, "macro", "M", nil, "XLM macro `code`")
	dcomExcelMacroExecFlags.Flags.StringVar(&dcomExcelMacro.MacroFile, "macro-file", "", "XLM macro `file`")
	registerExecutionFlags(dcomExcelMacroExecFlags.Flags)
	registerExecutionOutputFlags(dcomExcelMacroExecFlags.Flags)
	registerExecutionUploadFlags(dcomExcelMacroExecFlags.Flags)

	cmdFlags[dcomExcelMacroCmd] = []*flagSet{
		dcomExcelMacroExecFlags,
		defaultAuthFlags,
		defaultLogFlags,
		defaultNetRpcFlags,
	}
	dcomExcelMacroCmd.Flags().AddFlagSet(dcomExcelMacroExecFlags.Flags)

	// Constraints
	dcomExcelMacroCmd.MarkFlagsOneRequired("command", "exec", "macro", "macro-file", "upload")
	dcomExcelMacroCmd.MarkFlagsMutuallyExclusive("command", "exec", "macro", "macro-file")
	dcomExcelMacroCmd.MarkFlagsMutuallyExclusive("macro", "macro-file", "out")
}

func dcomVisualStudioDteCmdInit() {
	dcomVisualStudioDteVsFlags := newFlagSet("Visual Studio")
	dcomVisualStudioDteVsFlags.Flags.BoolVar(&dcomVisualStudioDte.Is2019, "vs-2019", false, "Target Visual Studio 2019")
	dcomVisualStudioDteVsFlags.Flags.StringVar(&dcomVisualStudioDte.CommandName, "vs-command", "", "Visual Studio DTE command to execute")
	dcomVisualStudioDteVsFlags.Flags.StringVar(&dcomVisualStudioDte.CommandArgs, "vs-args", "", "Visual Studio DTE command arguments")

	dcomVisualStudioDteExecFlags := newFlagSet("Execution")
	registerExecutionFlags(dcomVisualStudioDteExecFlags.Flags)
	registerExecutionOutputFlags(dcomVisualStudioDteExecFlags.Flags)
	registerExecutionUploadFlags(dcomVisualStudioDteExecFlags.Flags)

	cmdFlags[dcomVisualStudioDteCmd] = []*flagSet{
		dcomVisualStudioDteVsFlags,
		dcomVisualStudioDteExecFlags,
		defaultAuthFlags,
		defaultLogFlags,
		defaultNetRpcFlags,
	}
	dcomVisualStudioDteCmd.Flags().AddFlagSet(dcomVisualStudioDteVsFlags.Flags)
	dcomVisualStudioDteCmd.Flags().AddFlagSet(dcomVisualStudioDteExecFlags.Flags)

	// Constraints
	dcomVisualStudioDteCmd.MarkFlagsOneRequired("command", "exec", "vs-command", "upload")
	dcomVisualStudioDteCmd.MarkFlagsMutuallyExclusive("command", "exec", "vs-command")
	dcomVisualStudioDteCmd.MarkFlagsMutuallyExclusive("vs-command", "out")
}

func dcomExcelXllCmdInit() {
  dcomExcelXllExecFlags := newFlagSet("Execution")
  dcomExcelXllExecFlags.Flags.StringVar(&dcomExcelXll.XllLocation, "xll", "", "XLL/DLL local or UNC `path`")

  cmdFlags[dcomExcelXllCmd] = []*flagSet{
    dcomExcelXllExecFlags,
    defaultAuthFlags,
    defaultLogFlags,
    defaultNetRpcFlags,
  }
  dcomExcelXllCmd.Flags().AddFlagSet(dcomExcelXllExecFlags.Flags)

  // Constraints
  if err := dcomExcelXllCmd.MarkFlagRequired("xll"); err != nil {
    panic(err)
  }
}

var (
  dcomMmc                = dcomexec.DcomMmc{}
  dcomShellWindows       = dcomexec.DcomShellWindows{}
  dcomShellBrowserWindow = dcomexec.DcomShellBrowserWindow{}
  dcomHtafile            = dcomexec.DcomHtafile{}
  dcomExcelMacro         = dcomexec.DcomExcelMacro{}
  dcomExcelXll           = dcomexec.DcomExcelXll{}
  dcomVisualStudioDte    = dcomexec.DcomVisualStudioDte{}

  dcomCmd = &cobra.Command{
    Use:   "dcom",
    Short: "Execute with Distributed Component Object Model (MS-DCOM)",
    Long: `Description:
  The dcom module uses exposed Distributed Component Object Model (DCOM) objects to spawn processes.`,
    GroupID: "module",
    Args:    cobra.ArbitraryArgs,
  }

  dcomExcelCmd = &cobra.Command{
    Use:   "excel [method]",
    Short: "Execute with DCOM object(s) targeting Microsoft Excel",
    Long: `Description:
  Commands in the excel group use exposed Excel DCOM objects to gain remote execution`,
  }

  dcomVisualStudioCmd = &cobra.Command{
    Use:   "visualstudio [method]",
    Short: "Execute with DCOM object(s) targeting Microsoft Visual Studio",
    Long: `Description:
  Commands in the visualstudio group use exposed Visual Studio DCOM objects to gain remote execution`,
  }

  dcomMmcCmd = &cobra.Command{
    Use:   "mmc [target]",
    Short: "Execute with the MMC20.Application DCOM object",
    Long: `Description:
  The mmc method uses the exposed MMC20.Application object to call Document.ActiveView.ShellExec,
  and ultimately spawn a process on the remote host.`,
		Args: args(argsRpcClient("cifs", ""),
			argsOutput("smb"),
			argsUpload("smb"),
			argsAcceptValues("window", &dcomMmc.WindowState, "Minimized", "Maximized", "Restored"),
		),
		Run: func(cmd *cobra.Command, args []string) {
			dcomMmc.Client = &rpcClient
      ctx := log.With().Str("module", dcomexec.ModuleName).Str("method", dcomexec.MethodMmc).
        Logger().WithContext(gssapi.NewSecurityContext(context.Background()))

      if err := goexec.ExecuteCleanMethod(ctx, &dcomMmc, &exec); err != nil {
        log.Fatal().Err(err).Msg("Operation failed")
      }
    },
  }

  dcomShellWindowsCmd = &cobra.Command{
    Use:   "shellwindows [target]",
    Short: "Execute with the ShellWindows DCOM object",
    Long: `Description:
  The shellwindows method uses the exposed ShellWindows DCOM object on older Windows installations
  to call Item().Document.Application.ShellExecute, and spawn the provided process.`,
		Args: args(argsRpcClient("host", ""),
			argsOutput("smb"),
			argsUpload("smb"),
			argsAcceptValues("app-window", &dcomShellWindows.WindowState, "0", "1", "2", "3", "4", "5", "7", "10"),
		),
		Run: func(cmd *cobra.Command, args []string) {
			dcomShellWindows.Client = &rpcClient
      ctx := log.With().Str("module", dcomexec.ModuleName).Str("method", dcomexec.MethodShellWindows).
        Logger().WithContext(gssapi.NewSecurityContext(context.Background()))

      if err := goexec.ExecuteCleanMethod(ctx, &dcomShellWindows, &exec); err != nil {
        log.Fatal().Err(err).Msg("Operation failed")
      }
    },
  }

  dcomShellBrowserWindowCmd = &cobra.Command{
    Use:   "shellbrowserwindow [target]",
    Short: "Execute with the ShellBrowserWindow DCOM object",
    Long: `Description:
  The shellbrowserwindow method uses the exposed ShellBrowserWindow DCOM object on older Windows installations
  to call Document.Application.ShellExecute, and spawn the provided process.`,
		Args: args(argsRpcClient("host", ""),
			argsOutput("smb"),
			argsUpload("smb"),
			argsAcceptValues("app-window", &dcomShellBrowserWindow.WindowState, "0", "1", "2", "3", "4", "5", "7", "10"),
		),
		Run: func(cmd *cobra.Command, args []string) {
			dcomShellBrowserWindow.Client = &rpcClient
      ctx := log.With().Str("module", dcomexec.ModuleName).Str("method", dcomexec.MethodShellBrowserWindow).
        Logger().WithContext(gssapi.NewSecurityContext(context.Background()))

      if err := goexec.ExecuteCleanMethod(ctx, &dcomShellBrowserWindow, &exec); err != nil {
        log.Fatal().Err(err).Msg("Operation failed")
      }
    },
  }

  dcomHtafileCmd = &cobra.Command{
    Use:   "htafile [target]",
    Short: "Execute with the HTAFile DCOM object",
    Long: `Description:
  The htafile method uses the exposed "HTML Application" DCOM object to load a remote HTA application or execute inline.
  This is made possible by the Load method of the IPersistMoniker interface.`,
		Args: args(argsRpcClient("host", ""), argsOutput("smb"), argsUpload("smb")),
		RunE: func(cmd *cobra.Command, args []string) error {
			dcomHtafile.Client = &rpcClient
      dcomHtafile.Url = dcomexec.HtafileGetUrl(dcomHtafile.Url, dcomHtafile.Javascript, dcomHtafile.Vbscript, &exec)

      if url := strings.ToLower(dcomHtafile.Url); (strings.HasPrefix(url, "javascript:") || strings.HasPrefix(url, "vbscript:")) && len(url) > 508 {
        return fmt.Errorf("script URL exceeds maximum length supported by mshta.exe (%d > 508)", len(url))
      }
      ctx := log.With().Str("module", dcomexec.ModuleName).Str("method", dcomexec.MethodHtafile).
        Logger().WithContext(gssapi.NewSecurityContext(context.Background()))

      if err := goexec.ExecuteCleanMethod(ctx, &dcomHtafile, &exec); err != nil {
        log.Fatal().Err(err).Msg("Operation failed")
      }
      return nil
    },
  }

  dcomExcelMacroCmd = &cobra.Command{
    Use:   "macro [target]",
    Short: "Execute using Excel 4.0 macros (XLM)",
    Long: `Description:
  The macro method uses the exposed Excel.Application DCOM object to call ExecuteExcel4Macro, thus executing
  XLM macros at will. This method requires that the remote host has Microsoft Excel installed.`,
		Args: args(argsRpcClient("host", ""), argsOutput("smb"), argsUpload("smb"),
			func(*cobra.Command, []string) error {
				if dcomExcelMacro.MacroFile != "" {
          f, err := os.Open(dcomExcelMacro.MacroFile)
          if err != nil {
            return fmt.Errorf("open macro file: %w", err)
          }
          defer func() { _ = f.Close() }()
          b, err := io.ReadAll(f)
          if err != nil {
            return fmt.Errorf("read macro file: %w", err)
          }
          dcomExcelMacro.Macros = strings.Split(string(b), "\n")
        }
        return nil
      },
    ),
    Run: func(*cobra.Command, []string) {
      dcomExcelMacro.Client = &rpcClient
      ctx := log.With().Str("module", dcomexec.ModuleName).Str("method", dcomexec.MethodExcelMacro).
        Logger().WithContext(gssapi.NewSecurityContext(context.Background()))

      if err := goexec.ExecuteCleanMethod(ctx, &dcomExcelMacro, &exec); err != nil {
        log.Fatal().Err(err).Msg("Operation failed")
      }
    },
  }

  dcomExcelXllCmd = &cobra.Command{
    Use:   "xll [target]",
    Short: "Execute by Loading an XLL add-in",
    Long: `Description:
  The xll method uses the exposed Excel.Application DCOM object to call RegisterXLL, thus loading a XLL/DLL.
  The XLL location (--xll) can be a path on the remote filesystem or an UNC path. This method requires that the
  remote host has Microsoft Excel installed.`,
    Args: args(argsRpcClient("host", "")),
    Run: func(*cobra.Command, []string) {
      dcomExcelXll.Client = &rpcClient
      ctx := log.With().Str("module", dcomexec.ModuleName).Str("method", dcomexec.MethodExcelXLL).
        Logger().WithContext(gssapi.NewSecurityContext(context.Background()))

      if err := goexec.ExecuteCleanAuxiliaryMethod(ctx, &dcomExcelXll); err != nil {
        log.Fatal().Err(err).Msg("Operation failed")
      }
    },
  }

  dcomVisualStudioDteCmd = &cobra.Command{
    Use:   "dte [target]",
    Short: "Execute with the VisualStudio.DTE object",
    Long: `Description:
  The dte method uses the exposed VisualStudio.DTE object to spawn a process via the ExecuteCommand method. This method
  requires that the remote host has Microsoft Visual Studio installed.`,
		Args: args(argsRpcClient("host", ""), argsOutput("smb"), argsUpload("smb")),
		Run: func(*cobra.Command, []string) {
			dcomVisualStudioDte.Client = &rpcClient
      ctx := log.With().Str("module", dcomexec.ModuleName).Str("method", dcomexec.MethodVisualStudioDTE).
        Logger().WithContext(gssapi.NewSecurityContext(context.Background()))

      if err := goexec.ExecuteCleanMethod(ctx, &dcomVisualStudioDte, &exec); err != nil {
        log.Fatal().Err(err).Msg("Operation failed")
      }
    },
  }
)
