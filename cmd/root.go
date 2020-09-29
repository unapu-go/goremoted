// Copyright © 2020 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/moisespsena-go/logging"
	path_helpers "github.com/moisespsena-go/path-helpers"
	"github.com/pkg/errors"

	"github.com/fsnotify/fsnotify"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/unapu-go/goremoted/internal"
)

var (
	cfgFile string
	configs = make(chan *internal.Config)
	log     = logging.GetOrCreateLogger(path_helpers.GetCalledDir())
	Viper   = viper.NewWithOptions(viper.KeyDelimiter("\\"))
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "goremoted",
	Short: "The Golang projects remote server bridge",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		go onConfig()
		return internal.Run(configs)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml or $HOME/.goremoted.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		Viper.SetConfigFile(cfgFile)
	} else if fileExists("config") {
		// Search config in ./ directory with name "config" (without extension).
		Viper.AddConfigPath(".")
		Viper.SetConfigName("config")
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".goremoted" (without extension).
		Viper.AddConfigPath(home)
		Viper.SetConfigName(".goremoted")
	}

	Viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := Viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", Viper.ConfigFileUsed())
	}

	Viper.WatchConfig()
	Viper.OnConfigChange(func(in fsnotify.Event) {
		onConfig()
	})
}

func fileExists(pth ...string) bool {
	for _, p := range pth {
		for _, ext := range []string{"yaml", "yml", "toml", "json"} {
			if s, err := os.Stat(p + "." + ext); err == nil && !s.IsDir() {
				return true
			}
		}
	}
	return false
}

func onConfig() {
	var cfg internal.Config
	if err := Viper.Unmarshal(&cfg); err == nil {
		configs <- &cfg
	} else {
		log.Error(errors.Wrap(err, "parse config"))
	}
}
