// Copyright © 2016 Prateek Malhotra (someone1@gmail.com)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"context"
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dustin/go-humanize"
	_ "github.com/joho/godotenv/autoload"
	"github.com/juju/ratelimit"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/someone1/zfsbackup-go/config"
	"github.com/someone1/zfsbackup-go/files"
	"github.com/someone1/zfsbackup-go/zfs"
)

var (
	numCores         int
	logLevel         string
	workingDirectory string
	errInvalidInput  = errors.New("invalid input")
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "zfsbackup",
	Short: "zfsbackup is a tool used to do off-site backups of ZFS volumes.",
	Long: `zfsbackup is a tool used to do off-site backups of ZFS volumes.
It leverages the built-in snapshot capabilities of ZFS in order to export ZFS
volumes for long-term storage.

zfsbackup uses the "zfs send" command to export, and optionally compress, sign,
encrypt, and split the send stream to files that are then transferred to a
destination of your choosing.`,
	PersistentPreRunE: processFlags,
	PersistentPostRun: postRunCleanup,
	SilenceErrors:     true,
	SilenceUsage:      true,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ctx context.Context) {
	if err := RootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(func() {
		zapConfig := zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeCaller = nil
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

		if logLevel, err := zap.ParseAtomicLevel(logLevel); err != nil {
			zapConfig.Level = logLevel
		}

		zapLogger, _ := zapConfig.Build(zap.AddStacktrace(zap.DPanicLevel))

		zap.ReplaceGlobals(zapLogger)
	})
	cobra.OnFinalize(func() {
		_ = zap.L().Sync()
	})

	RootCmd.PersistentFlags().IntVar(
		&numCores,
		"numCores",
		2,
		"number of CPU cores to utilize. Do not exceed the number of CPU cores on the system.",
	)
	RootCmd.PersistentFlags().StringVar(
		&logLevel,
		"logLevel",
		"info",
		"this controls the verbosity level of logging. Possible values are critical, error, warning, info, debug.",
	)
	RootCmd.PersistentFlags().StringVar(
		&workingDirectory,
		"workingDirectory",
		"~/.zfsbackup",
		"the working directory path for zfsbackup.",
	)
	RootCmd.PersistentFlags().StringVar(
		&jobInfo.ManifestPrefix,
		"manifestPrefix",
		"manifests", "the prefix to use for all manifest files.",
	)
	RootCmd.PersistentFlags().StringVar(
		&jobInfo.AesEncryptionKey,
		"encryptionKey",
		"",
		"aes encryption key used to encrypt/decrypt files (or use `ENCRYPTION_KEY` environment variable).)",
	)
	RootCmd.PersistentFlags().StringVar(
		&zfs.ZFSPath,
		"zfsPath",
		"zfs",
		"the path to the zfs executable.",
	)
	RootCmd.PersistentFlags().BoolVar(
		&config.JSONOutput,
		"jsonOutput",
		false,
		"dump results as a JSON string - on success only",
	)
	_ = []byte(os.Getenv("PGP_PASSPHRASE"))
}

func resetRootFlags() {
	jobInfo = files.JobInfo{}
	numCores = 2
	logLevel = "notice"
	workingDirectory = "~/.zfsbackup"
	jobInfo.ManifestPrefix = "manifests"
	zfs.ZFSPath = "zfs"
	config.JSONOutput = false
}

// nolint:gocyclo,funlen // Will do later
func processFlags(cmd *cobra.Command, args []string) error {
	if numCores <= 0 {
		zap.S().Errorf("The number of cores to use provided is an invalid value. It must be greater than 0. %d was given.", numCores)
		return errInvalidInput
	}

	if numCores > runtime.NumCPU() {
		zap.S().Warnf(
			"Ignoring user provided number of cores (%d) and using the number of detected cores (%d).",
			numCores, runtime.NumCPU(),
		)
		numCores = runtime.NumCPU()
	}
	zap.S().Infof("Setting number of cores to: %d", numCores)
	runtime.GOMAXPROCS(numCores)

	if err := setupGlobalVars(); err != nil {
		return err
	}
	zap.S().Infof("Setting working directory to %s", workingDirectory)
	return nil
}

func postRunCleanup(cmd *cobra.Command, args []string) {
	err := os.RemoveAll(config.BackupTempdir)
	if err != nil {
		zap.S().Errorf("Could not clean working temporary directory - %v", err)
	}
}

// nolint:gocyclo,funlen // Will do later
func setupGlobalVars() error {
	// Setup Tempdir
	if strings.HasPrefix(workingDirectory, "~") {
		usr, err := user.Current()
		if err != nil {
			zap.S().Errorf("Could not get current user due to error - %v", err)
			return err
		}
		workingDirectory = filepath.Join(usr.HomeDir, strings.TrimPrefix(workingDirectory, "~"))
	}

	if dir, serr := os.Stat(workingDirectory); serr == nil && !dir.IsDir() {
		zap.S().Errorf(
			"Cannot create working directory because another non-directory object already exists in that path (%s)",
			workingDirectory,
		)
		return errInvalidInput
	} else if serr != nil {
		err := os.Mkdir(workingDirectory, 0755)
		if err != nil {
			zap.S().Errorf("Could not create working directory %s due to error - %v", workingDirectory, err)
			return err
		}
	}

	if len(jobInfo.AesEncryptionKey) == 0 {
		jobInfo.AesEncryptionKey = os.Getenv("ENCRYPTION_KEY")
	}

	dirPath := filepath.Join(workingDirectory, "temp")
	if dir, serr := os.Stat(dirPath); serr == nil && !dir.IsDir() {
		zap.S().Errorf(
			"Cannot create temp dir in working directory because another non-directory object already exists in that path (%s)",
			dirPath,
		)
		return errInvalidInput
	} else if serr != nil {
		err := os.Mkdir(dirPath, 0755)
		if err != nil {
			zap.S().Errorf("Could not create temp directory %s due to error - %v", dirPath, err)
			return err
		}
	}

	tempdir, err := os.MkdirTemp(dirPath, config.ProgramName)
	if err != nil {
		zap.S().Errorf("Could not create temp directory due to error - %v", err)
		return err
	}

	config.BackupTempdir = tempdir
	config.WorkingDir = workingDirectory

	dirPath = filepath.Join(workingDirectory, "cache")
	if dir, serr := os.Stat(dirPath); serr == nil && !dir.IsDir() {
		zap.S().Errorf(
			"Cannot create cache dir in working directory because another non-directory object already exists in that path (%s)",
			dirPath,
		)
		return errInvalidInput
	} else if serr != nil {
		err := os.Mkdir(dirPath, 0755)
		if err != nil {
			zap.S().Errorf("Could not create cache directory %s due to error - %v", dirPath, err)
			return err
		}
	}

	if maxUploadSpeed != 0 {
		zap.S().Infof("Limiting the upload speed to %s/s.", humanize.Bytes(maxUploadSpeed*humanize.KByte))
		config.BackupUploadBucket = ratelimit.NewBucketWithRate(float64(maxUploadSpeed*humanize.KByte), int64(maxUploadSpeed*humanize.KByte))
	}
	return nil
}
