package cmd

import (
  "context"
  "fmt"
  "time"

  "github.com/FalconOpsLLC/goexec/internal/util"
  "github.com/FalconOpsLLC/goexec/pkg/goexec"
  tschexec "github.com/FalconOpsLLC/goexec/pkg/goexec/tsch"
  "github.com/oiweiwei/go-msrpc/ssp/gssapi"
  "github.com/spf13/cobra"
)

func tschCmdInit() {
  cmdFlags[tschCmd] = []*flagSet{
    defaultAuthFlags,
    defaultLogFlags,
    defaultNetRpcFlags,
  }
  tschDemandCmdInit()
  tschCreateCmdInit()
  tschChangeCmdInit()

  tschCmd.PersistentFlags().AddFlagSet(defaultAuthFlags.Flags)
  tschCmd.PersistentFlags().AddFlagSet(defaultLogFlags.Flags)
  tschCmd.PersistentFlags().AddFlagSet(defaultNetRpcFlags.Flags)
  tschCmd.AddCommand(tschDemandCmd, tschCreateCmd, tschChangeCmd)
}

func tschDemandCmdInit() {
  tschDemandFlags := newFlagSet("Task Scheduler")

  tschDemandFlags.Flags.StringVarP(&tschTask, "task", "t", "", "Name or path of the new task")
  tschDemandFlags.Flags.Uint32Var(&tschDemand.SessionId, "session", 0, "Hijack existing session given the session `ID`")
  tschDemandFlags.Flags.StringVar(&tschDemand.UserSid, "sid", "S-1-5-18", "User `SID` to impersonate")
  tschDemandFlags.Flags.BoolVar(&tschDemand.NoDelete, "no-delete", false, "Don't delete task after execution")

	tschDemandExecFlags := newFlagSet("Execution")

	registerExecutionFlags(tschDemandExecFlags.Flags)
	registerExecutionOutputFlags(tschDemandExecFlags.Flags)
	registerExecutionUploadFlags(tschDemandExecFlags.Flags)

	cmdFlags[tschDemandCmd] = []*flagSet{
    tschDemandFlags,
    tschDemandExecFlags,
    defaultAuthFlags,
    defaultLogFlags,
    defaultNetRpcFlags,
  }

	tschDemandCmd.Flags().AddFlagSet(tschDemandFlags.Flags)
	tschDemandCmd.Flags().AddFlagSet(tschDemandExecFlags.Flags)
	tschDemandCmd.MarkFlagsOneRequired("exec", "command", "upload")
}

func tschCreateCmdInit() {
  tschCreateFlags := newFlagSet("Task Scheduler")

  tschCreateFlags.Flags.StringVarP(&tschTask, "task", "t", "", "Name or path of the new task")
  tschCreateFlags.Flags.DurationVar(&tschCreate.StopDelay, "delay-stop", 5*time.Second, "Delay between task execution and termination. This won't stop the spawned process")
  tschCreateFlags.Flags.DurationVar(&tschCreate.StartDelay, "start-delay", 5*time.Second, "Delay between task registration and execution")
  //tschCreateFlags.Flags.DurationVar(&tschCreate.DeleteDelay, "delete-delay", 0*time.Second, "Delay between task termination and deletion")
  tschCreateFlags.Flags.BoolVar(&tschCreate.NoDelete, "no-delete", false, "Don't delete task after execution")
  tschCreateFlags.Flags.BoolVar(&tschCreate.CallDelete, "call-delete", false, "Directly call SchRpcDelete to delete task")
  tschCreateFlags.Flags.StringVar(&tschCreate.UserSid, "sid", "S-1-5-18", "User `SID` to impersonate")

	tschCreateExecFlags := newFlagSet("Execution")

	registerExecutionFlags(tschCreateExecFlags.Flags)
	registerExecutionOutputFlags(tschCreateExecFlags.Flags)
	registerExecutionUploadFlags(tschCreateExecFlags.Flags)

	cmdFlags[tschCreateCmd] = []*flagSet{
    tschCreateFlags,
    tschCreateExecFlags,
    defaultAuthFlags,
    defaultLogFlags,
    defaultNetRpcFlags,
  }

	tschCreateCmd.Flags().AddFlagSet(tschCreateFlags.Flags)
	tschCreateCmd.Flags().AddFlagSet(tschCreateExecFlags.Flags)
	tschCreateCmd.MarkFlagsOneRequired("exec", "command", "upload")
}

func tschChangeCmdInit() {
  tschChangeFlags := newFlagSet("Task Scheduler")

  tschChangeFlags.Flags.StringVarP(&tschChange.TaskPath, "task", "t", "", "Path to existing task")
  tschChangeFlags.Flags.BoolVar(&tschChange.NoStart, "no-start", false, "Don't start the task")
  tschChangeFlags.Flags.BoolVar(&tschChange.NoRevert, "no-revert", false, "Don't restore the original task definition")

	tschChangeExecFlags := newFlagSet("Execution")

	registerExecutionFlags(tschChangeExecFlags.Flags)
	registerExecutionOutputFlags(tschChangeExecFlags.Flags)
	registerExecutionUploadFlags(tschChangeExecFlags.Flags)

	cmdFlags[tschChangeCmd] = []*flagSet{
    tschChangeFlags,
    tschChangeExecFlags,
    defaultAuthFlags,
    defaultLogFlags,
    defaultNetRpcFlags,
  }

  tschChangeCmd.Flags().AddFlagSet(tschChangeFlags.Flags)
  tschChangeCmd.Flags().AddFlagSet(tschChangeExecFlags.Flags)

  // Constraints
  {
		if err := tschChangeCmd.MarkFlagRequired("task"); err != nil {
			panic(err)
		}
		tschChangeCmd.MarkFlagsOneRequired("exec", "command", "upload")
  }
}

func argsTask(*cobra.Command, []string) error {
  switch {
  case tschTask == "":
    tschTask = `\` + util.RandomString()
  case tschexec.ValidateTaskPath(tschTask) == nil:
  case tschexec.ValidateTaskName(tschTask) == nil:
    tschTask = `\` + tschTask
  default:
    return fmt.Errorf("invalid task Label or path: %q", tschTask)
  }
  return nil
}

var (
  tschDemand tschexec.TschDemand
  tschCreate tschexec.TschCreate
  tschChange tschexec.TschChange

  tschTask string

  tschCmd = &cobra.Command{
    Use:   "tsch",
    Short: "Execute with Windows Task Scheduler (MS-TSCH)",
    Long: `Description:
  The tsch module makes use of the Windows Task Scheduler service (MS-TSCH) to
  spawn processes on the remote target.`,
    GroupID: "module",
    Args:    cobra.NoArgs,
  }

  tschDemandCmd = &cobra.Command{
    Use:   "demand [target]",
    Short: "Register a remote scheduled task and demand immediate start",
    Long: `Description:
  Similar to the create method, the demand method will call SchRpcRegisterTask,
  But rather than setting a defined time when the task will start, it will
  additionally call SchRpcRun to forcefully start the task.`,
		Args: args(
			argsRpcClient("cifs", "ncacn_np:[atsvc]"),
			argsOutput("smb"),
			argsUpload("smb"),
			argsTask,
		),

		Run: func(*cobra.Command, []string) {
			tschDemand.Client = &rpcClient
      tschDemand.TaskPath = tschTask

      ctx := log.With().
        Str("module", "tsch").
        Str("method", "demand").
        Logger().WithContext(gssapi.NewSecurityContext(context.TODO()))

      if err := goexec.ExecuteCleanMethod(ctx, &tschDemand, &exec); err != nil {
        log.Fatal().Err(err).Msg("Operation failed")
      }
    },
  }
  tschCreateCmd = &cobra.Command{
    Use:   "create [target]",
    Short: "Create a remote scheduled task with an automatic start time",
    Long: `Description:
  The create method calls SchRpcRegisterTask to register a scheduled task
  with an automatic start time.This method avoids directly calling SchRpcRun,
  and can even avoid calling SchRpcDelete by populating the DeleteExpiredTaskAfter
  Setting.`,
		Args: args(
			argsRpcClient("cifs", "ncacn_np:[atsvc]"),
			argsOutput("smb"),
			argsUpload("smb"),
			argsTask,
		),

		Run: func(*cobra.Command, []string) {
			tschCreate.Client = &rpcClient
      tschCreate.TaskPath = tschTask

      ctx := log.With().
        Str("module", "tsch").
        Str("method", "create").
        Logger().WithContext(gssapi.NewSecurityContext(context.TODO()))

      if err := goexec.ExecuteCleanMethod(ctx, &tschCreate, &exec); err != nil {
        log.Fatal().Err(err).Msg("Operation failed")
      }
    },
  }
  tschChangeCmd = &cobra.Command{
    Use:   "change [target]",
    Short: "Modify an existing task to spawn an arbitrary process",
    Long: `Description:
  The change method calls SchRpcRetrieveTask to fetch the definition of an existing
  task (-t), then modifies the task definition to spawn a process`,
		Args: args(
			argsRpcClient("cifs", "ncacn_np:[atsvc]"),
			argsOutput("smb"),
			argsUpload("smb"),

			func(*cobra.Command, []string) error {
				return tschexec.ValidateTaskPath(tschChange.TaskPath)
			},
		),

    Run: func(*cobra.Command, []string) {
      tschChange.Client = &rpcClient
      tschChange.IO = exec

      ctx := log.With().
        Str("module", "tsch").
        Str("method", "change").
        Logger().WithContext(gssapi.NewSecurityContext(context.TODO()))

      if err := goexec.ExecuteCleanMethod(ctx, &tschChange, &exec); err != nil {
        log.Fatal().Err(err).Msg("Operation failed")
      }
    },
  }
)
