package cmd

import (
  "context"

  "github.com/FalconOpsLLC/goexec/internal/util"
  "github.com/FalconOpsLLC/goexec/pkg/goexec"
  "github.com/oiweiwei/go-msrpc/ssp/gssapi"
  "github.com/spf13/cobra"

  scmrexec "github.com/FalconOpsLLC/goexec/pkg/goexec/scmr"
)

func scmrCmdInit() {
  cmdFlags[scmrCmd] = []*flagSet{
    defaultAuthFlags,
    defaultLogFlags,
    defaultNetRpcFlags,
  }
  scmrCreateCmdInit()
  scmrChangeCmdInit()
  scmrDeleteCmdInit()

  scmrCmd.PersistentFlags().AddFlagSet(defaultAuthFlags.Flags)
  scmrCmd.PersistentFlags().AddFlagSet(defaultLogFlags.Flags)
  scmrCmd.PersistentFlags().AddFlagSet(defaultNetRpcFlags.Flags)
  scmrCmd.AddCommand(scmrCreateCmd, scmrChangeCmd, scmrDeleteCmd)
}

func scmrCreateCmdInit() {
	scmrCreateFlags := newFlagSet("Service")

	scmrCreateFlags.Flags.StringVarP(&scmrCreate.DisplayName, "display-name", "n", "", "Display name of service to create")
	scmrCreateFlags.Flags.StringVarP(&scmrCreate.ServiceName, "service", "s", "", "Name of service to create")
	scmrCreateFlags.Flags.BoolVar(&scmrCreate.NoDelete, "no-delete", false, "Don't delete service after execution")
	scmrCreateFlags.Flags.BoolVar(&scmrCreate.NoStart, "no-start", false, "Don't start service")

	scmrCreateExecFlags := newFlagSet("Execution")

	// TODO: SCMR output
	//registerExecutionOutputFlags(scmrCreateExecFlags.Flags)
	registerExecutionUploadFlags(scmrCreateExecFlags.Flags)

	scmrCreateExecFlags.Flags.StringVarP(&exec.Input.ExecutablePath, "executable-path", "f", "", "Full path to a remote Windows executable")
	scmrCreateExecFlags.Flags.StringVarP(&exec.Input.Arguments, "args", "a", "", "Arguments to pass to the executable")

	scmrCreateCmd.Flags().AddFlagSet(scmrCreateFlags.Flags)
	scmrCreateCmd.Flags().AddFlagSet(scmrCreateExecFlags.Flags)

	cmdFlags[scmrCreateCmd] = []*flagSet{
		scmrCreateExecFlags,
		scmrCreateFlags,
		defaultAuthFlags,
		defaultLogFlags,
		defaultNetRpcFlags,
	}

	// Constraints
	{
		//scmrCreateCmd.MarkFlagsMutuallyExclusive("no-delete", "no-start")
		scmrCreateCmd.MarkFlagsOneRequired("executable-path", "upload")
	}
}

func scmrChangeCmdInit() {
	scmrChangeFlags := newFlagSet("Service Control")

	scmrChangeFlags.Flags.StringVarP(&scmrChange.ServiceName, "service-name", "s", "", "Name of service to modify")
	scmrChangeFlags.Flags.BoolVar(&scmrChange.NoStart, "no-start", false, "Don't start service")

	scmrChangeExecFlags := newFlagSet("Execution")

	scmrChangeExecFlags.Flags.StringVarP(&exec.Input.ExecutablePath, "executable-path", "f", "", "Full path to remote Windows executable")
	scmrChangeExecFlags.Flags.StringVarP(&exec.Input.Arguments, "args", "a", "", "Arguments to pass to executable")

	// TODO: SCMR output
	//registerExecutionOutputFlags(scmrChangeExecFlags.Flags)
	registerExecutionUploadFlags(scmrChangeExecFlags.Flags)

	cmdFlags[scmrChangeCmd] = []*flagSet{
		scmrChangeFlags,
		scmrChangeExecFlags,
		defaultAuthFlags,
		defaultLogFlags,
		defaultNetRpcFlags,
	}

	scmrChangeCmd.Flags().AddFlagSet(scmrChangeFlags.Flags)
	scmrChangeCmd.Flags().AddFlagSet(scmrChangeExecFlags.Flags)

	// Constraints
	{
		if err := scmrChangeCmd.MarkFlagRequired("service-name"); err != nil {
			panic(err)
		}
		scmrChangeCmd.MarkFlagsOneRequired("executable-path", "upload")
	}
}

func scmrDeleteCmdInit() {
  scmrDeleteFlags := newFlagSet("Service Control")
  scmrDeleteFlags.Flags.StringVarP(&scmrDelete.ServiceName, "service-name", "s", scmrDelete.ServiceName, "Name of service to delete")

  cmdFlags[scmrDeleteCmd] = []*flagSet{
    scmrDeleteFlags,
    defaultAuthFlags,
    defaultLogFlags,
    defaultNetRpcFlags,
  }

  scmrDeleteCmd.Flags().AddFlagSet(scmrDeleteFlags.Flags)

  if err := scmrDeleteCmd.MarkFlagRequired("service-name"); err != nil {
    panic(err)
  }
}

var (
  scmrCreate = scmrexec.ScmrCreate{}
  scmrChange = scmrexec.ScmrChange{}
  scmrDelete = scmrexec.ScmrDelete{}

  scmrCmd = &cobra.Command{
    Use:   "scmr",
    Short: "Execute with Service Control Manager Remote (MS-SCMR)",
    Long: `Description:
  The SCMR module works a lot like Impacket's smbexec.py, but it provides additional RPC transports
  to evade network monitoring or firewall rules, and some minor OPSEC improvements overall.`,
    GroupID: "module",
    Args:    cobra.NoArgs,
  }

  scmrCreateCmd = &cobra.Command{
    Use:   "create [target]",
    Short: "Spawn a remote process by creating & running a Windows service",
    Long: `Description:
  The create method calls RCreateServiceW to create a new Windows service on the
  remote target with the provided executable & arguments as the lpBinaryPathName`,
		Args: args(
			argsRpcClient("cifs", "ncacn_np:[svcctl]"),
			argsSmbClient(),
			argsUpload("smb"),
		),

		Run: func(cmd *cobra.Command, args []string) {
			scmrCreate.Client = &rpcClient
      scmrCreate.IO = exec

      log = log.With().
        Str("module", "scmr").
        Str("method", "create").
        Logger()

      // Warnings
      {
        if scmrCreate.ServiceName == "" {
          log.Warn().Msg("No service name was provided. Using a random string")
          scmrCreate.ServiceName = util.RandomString()
        }
        if scmrCreate.NoDelete {
          log.Warn().Msg("Service will not be deleted after execution")
        }
        if scmrCreate.DisplayName == "" {
          log.Debug().Msg("No display name specified, using service name as display name")
          scmrCreate.DisplayName = scmrCreate.ServiceName
        }
      }

      ctx := log.WithContext(gssapi.NewSecurityContext(context.Background()))

      if err := goexec.ExecuteCleanMethod(ctx, &scmrCreate, &exec); err != nil {
        log.Fatal().Err(err).Msg("Operation failed")
      }
    },
  }

  scmrChangeCmd = &cobra.Command{
    Use:   "change [target]",
    Short: "Change an existing Windows service to spawn an arbitrary process",
    Long: `Description:
  The change method executes programs by modifying existing Windows services
  using the RChangeServiceConfigW method rather than calling RCreateServiceW
  like scmr create. The modified service is restored to its original state
  after execution`,
		Args: args(
			argsRpcClient("cifs", "ncacn_np:[svcctl]"),
			argsUpload("smb"),
		),

		Run: func(cmd *cobra.Command, args []string) {
			scmrChange.Client = &rpcClient
      scmrChange.IO = exec

      ctx := log.With().
        Str("module", "scmr").
        Str("method", "change").
        Logger().WithContext(gssapi.NewSecurityContext(context.Background()))

      if err := goexec.ExecuteCleanMethod(ctx, &scmrChange, &exec); err != nil {
        log.Fatal().Err(err).Msg("Operation failed")
      }
    },
  }
  scmrDeleteCmd = &cobra.Command{
    Use:   "delete [target]",
    Short: "Delete an existing Windows service",
    Long: `Description:
  The delete method will simply delete the provided service.`,

    Args: argsRpcClient("cifs", "ncacn_np:[svcctl]"),
    Run: func(cmd *cobra.Command, args []string) {
      scmrDelete.Client = &rpcClient

      ctx := log.With().
        Str("module", "scmr").
        Str("method", "delete").
        Logger().WithContext(gssapi.NewSecurityContext(context.Background()))

      if err := goexec.ExecuteCleanAuxiliaryMethod(ctx, &scmrDelete); err != nil {
        log.Fatal().Err(err).Msg("Operation failed")
      }
    },
  }
)
