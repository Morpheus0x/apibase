package cmd

import (
	"fmt"
	"reflect"

	"github.com/spf13/cobra"
	"gopkg.cc/apibase/log"
)

type CmdConfig struct {
	AppName           string
	ShortDescr        string
	LongDescr         string
	Version           string
	DefaultConfigPath string
	PrintFuct         PrintFuct
}

type PrintFuct func()

var defaultVersionPrint = func() {
	fmt.Printf("%s v%s\n", appConfig.AppName, appConfig.Version)
}

type Settings struct {
	ConfigFile  string
	ApiRoot     string
	Verbose     bool
	VeryVerbose bool
	Help        bool
}

func (s Settings) GetLogLevel() log.Level {
	if s.VeryVerbose {
		return log.LevelDebug
	}
	if s.Verbose {
		return log.LevelInfo
	}
	return log.LevelNotice
}

var (
	appConfig   CmdConfig
	appSettings Settings
	stopExec    bool = false
)

func ConfigureCLI(conf CmdConfig) *cobra.Command {
	appConfig = conf
	if reflect.ValueOf(conf.PrintFuct).IsZero() {
		appConfig.PrintFuct = defaultVersionPrint
	}
	return &cobra.Command{
		Use:   appConfig.AppName,
		Short: appConfig.ShortDescr,
		Long:  appConfig.LongDescr,
		Run: func(cmd *cobra.Command, args []string) {
			if cmd.Flag("version").Value.String() == "true" {
				appConfig.PrintFuct()
				stopExec = true
			}
		},
	}
}

// Parse cli arguments, returns true if program should exit
func Execute(root *cobra.Command) (Settings, bool) {
	defaultHelpFunc := root.HelpFunc()

	root.PersistentFlags().StringVarP(&appSettings.ConfigFile, "config", "c", appConfig.DefaultConfigPath, "config file")
	root.PersistentFlags().Bool("version", false, "print version, license and additional software info")
	root.PersistentFlags().BoolVarP(&appSettings.Verbose, "verbose", "v", false, "more detailed output, log level: info")
	root.PersistentFlags().BoolVar(&appSettings.VeryVerbose, "very-verbose", false, "even more detailed output, log level: debug")
	root.PersistentFlags().StringVarP(&appSettings.ApiRoot, "root", "r", "", "set api root behaviour, overwrites config file, specify a valid path for static file serving or an uri for reverse proxy")

	root.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		defaultHelpFunc(cmd, args)
		stopExec = true
	})

	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "print version, license and additional software info",
		Run: func(cmd *cobra.Command, args []string) {
			appConfig.PrintFuct()
			stopExec = true
		},
	})
	err := root.Execute()
	if err != nil {
		return appSettings, true
	}
	return appSettings, stopExec
}
