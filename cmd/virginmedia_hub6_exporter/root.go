package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	slogxtCobra "github.com/fornellas/slogxt/cobra"
	"github.com/fornellas/slogxt/log"
)

func getCmdChainStr(cmd *cobra.Command) string {
	cmdChain := []string{cmd.Name()}
	for {
		parentCmd := cmd.Parent()
		if parentCmd == nil {
			break
		}
		cmdChain = append([]string{parentCmd.Name()}, cmdChain...)
		cmd = parentCmd
	}
	return "⚙️ " + strings.Join(cmdChain, " ")
}

var RootCmd = &cobra.Command{
	Use:   "virginmedia_hub6_exporter",
	Short: "CLI G-Code Sender",
	Args:  cobra.NoArgs,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Environment Flags
		// Inspired by https://github.com/spf13/viper/issues/671#issuecomment-671067523
		v := viper.New()
		v.SetEnvPrefix("VM_HUB6_EXPORTER")
		v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
		v.AutomaticEnv()
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if !f.Changed && v.IsSet(f.Name) {
				cmd.Flags().Set(f.Name, fmt.Sprintf("%v", v.Get(f.Name)))
			}
		})

		// Logging
		logger := slogxtCobra.GetLogger(cmd.OutOrStderr()).
			WithGroup(getCmdChainStr(cmd))
		ctx := log.WithLogger(cmd.Context(), logger)
		cmd.SetContext(ctx)

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			logger := log.MustLogger(cmd.Context())
			logger.Error("Failed to display help", "err", err)
		}
		Exit(1)
	},
}

var resetFlagsFns = []func(){
	func() { slogxtCobra.Reset() },
}

func ResetFlags() {
	for _, resetFlagFn := range resetFlagsFns {
		resetFlagFn()
	}
}

func init() {
	slogxtCobra.AddLoggerFlags(RootCmd)

	resetFlagsFns = append(resetFlagsFns, func() {

	})
}
