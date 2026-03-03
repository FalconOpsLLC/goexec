package goexec

import (
  "context"
  "fmt"

  "github.com/rs/zerolog"
)

type Method interface {
  Connect(ctx context.Context) error
  Init(ctx context.Context) error
}

type CleanMethod interface {
  Method
  Clean
}

type ExecutionMethod interface {
  Method
  Execute(ctx context.Context, io *ExecutionIO) error
}

type CleanExecutionMethod interface {
  ExecutionMethod
  Clean
}

type AuxiliaryMethod interface {
  Method
  Call(ctx context.Context) error
}

type CleanAuxiliaryMethod interface {
  AuxiliaryMethod
  Clean
}

func ExecuteMethod(ctx context.Context, module ExecutionMethod, execIO *ExecutionIO) (err error) {
  log := zerolog.Ctx(ctx)

  if err = module.Connect(ctx); err != nil {
    log.Error().Err(err).Msg("Connection failed")
    return fmt.Errorf("connect: %w", err)
  }
  log.Debug().Msg("Module connected")

  if err = module.Init(ctx); err != nil {
    log.Error().Err(err).Msg("Module initialization failed")
    return fmt.Errorf("init module: %w", err)
  }
  log.Debug().Msg("Module initialized")

  if err = module.Execute(ctx, execIO); err != nil {
    log.Error().Err(err).Msg("Execution failed")
    return fmt.Errorf("execute: %w", err)
  }

  return
}

func ExecuteAuxiliaryMethod(ctx context.Context, module AuxiliaryMethod) (err error) {
  log := zerolog.Ctx(ctx)

  if err = module.Connect(ctx); err != nil {
    log.Error().Err(err).Msg("Connection failed")
    return fmt.Errorf("connect: %w", err)
  }
  log.Debug().Msg("Auxiliary module connected")

  if err = module.Init(ctx); err != nil {
    log.Error().Err(err).Msg("Module initialization failed")
    return fmt.Errorf("init module: %w", err)
  }
  log.Debug().Msg("Auxiliary module initialized")

  if err = module.Call(ctx); err != nil {
    log.Error().Err(err).Msg("Auxiliary method failed")
    return fmt.Errorf("call: %w", err)
  }
  log.Debug().Msg("Auxiliary method succeeded")

  return nil
}

func ExecuteCleanAuxiliaryMethod(ctx context.Context, module CleanAuxiliaryMethod) (err error) {
  log := zerolog.Ctx(ctx)

  defer func() {
    if err = module.Clean(ctx); err != nil {
      log.Error().Err(err).Msg("Module cleanup failed")
      err = nil
    }
  }()

  if err = ExecuteAuxiliaryMethod(ctx, module); err != nil {
    return fmt.Errorf("execute auxiliary method: %w", err)
  }
  return
}

func ExecuteCleanMethod(ctx context.Context, module CleanExecutionMethod, execIO *ExecutionIO) (err error) {
	log := zerolog.Ctx(ctx)

	// Connect
	if err = module.Connect(ctx); err != nil {
		log.Error().Err(err).Msg("Connection failed")
		return fmt.Errorf("connect: %w", err)
	}
	log.Debug().Msg("Module connected")

	// Init
	if err = module.Init(ctx); err != nil {
		log.Error().Err(err).Msg("Module initialization failed")
		return fmt.Errorf("init module: %w", err)
	}
	log.Debug().Msg("Module initialized")

	// Upload file (before execution)
	if execIO.Upload != nil && execIO.Upload.Provider != nil {
		log.Info().Str("dest", execIO.Upload.RemotePath).Msg("Uploading file")
		if err = execIO.DoUpload(ctx); err != nil {
			log.Error().Err(err).Msg("Upload failed")
			return fmt.Errorf("upload: %w", err)
		}
		log.Info().Msg("Upload succeeded")
	}

	// Execute (only if a command/executable was provided)
	if execIO.Input != nil && (execIO.Input.Executable != "" || execIO.Input.Command != "" || execIO.Input.ExecutablePath != "") {
		if err = module.Execute(ctx, execIO); err != nil {
			log.Error().Err(err).Msg("Execution failed")
			return fmt.Errorf("execute: %w", err)
		}
	}

	// Module cleanup
	if err = module.Clean(ctx); err != nil {
		log.Error().Err(err).Msg("Module cleanup failed")
		err = nil
	}

	// Output collection
	if execIO.Output != nil && execIO.Output.Provider != nil {
		log.Info().Msg("Collecting output")

		defer func() {
			if cleanErr := execIO.Clean(ctx); cleanErr != nil {
				log.Debug().Err(cleanErr).Msg("Output provider cleanup failed")
			}
		}()

		if err := execIO.GetOutput(ctx); err != nil {
			log.Error().Err(err).Msg("Output collection failed")
			return fmt.Errorf("get output: %w", err)
		}
		log.Debug().Msg("Output collection succeeded")
	}
	return
}
