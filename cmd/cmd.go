package cmd

import (
	"fmt"
	"os"

	"github.com/juju/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/risico/imagine/src/server"
)

var CLI struct {
	Start StartCMD `cmd:"" json:"start" help:"Starts the imagine application"`
}

type StartCMD struct {
	Hostname string `help:"sets the hostname for the server" json:"hostname"`
	Port     int    `help:"sets the application port" json:"port"`
}

func (r *StartCMD) Run() error {
	params := server.ServerParams{
		Hostname: r.Hostname,
		Port:     r.Port,
	}
	s := server.NewServer(params)
	return errors.Trace(s.Start())
}

var (
	cfgFile    string
	imagineCmd = &cobra.Command{
		Use:           "imagine",
		Short:         "imagine – imagine server",
		Long:          ``,
		Version:       "0.1.0",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
)

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigFile(".jamctl")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Config file used for imagine: ", viper.ConfigFileUsed())
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	imagineCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.jamctl.yaml)")
	viper.BindPFlag("config", imagineCmd.PersistentFlags().Lookup("config"))
}
