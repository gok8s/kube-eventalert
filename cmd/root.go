// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/gok8s/eventalert/utils/flog"

	"github.com/gok8s/eventalert/pkg"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/gok8s/eventalert/pkg/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "eventwatch",
	Short: "watch k8s event",
	Long:  `watch k8s event,filter and send alert`,
	Run: func(cmd *cobra.Command, args []string) {
		var config config.Config
		viper.Unmarshal(&config)

		switch config.LogLevel {
		case "info":
			flog.SetLevel(logrus.InfoLevel)
		case "debug":
			flog.SetLevel(logrus.DebugLevel)
		case "error":
			flog.SetLevel(logrus.ErrorLevel)
		case "warn":
			flog.SetLevel(logrus.WarnLevel)
		}

		logFile, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err == nil {
			mw := io.MultiWriter(os.Stdout, logFile)
			flog.SetOutput(mw)
		} else {
			flog.Log().Info("Failed to log to file, using default stderr")
		}

		defer logFile.Close()
		//TODO 必要参数的检查 和环境变量值的检查
		//controller.Start(config)
		pkg.Start(config)
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/ || conf/ ||/etc/eventwatch)")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	//box := packr.NewBox("../configs")
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err == nil {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// Search config in home directory with name ".eventwatch" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath("/etc/eventalert/configs")
		viper.AddConfigPath("/configs/")
		viper.AddConfigPath("./configs")
		viper.SetConfigType("yaml")

		config := os.Getenv("EALERT_CONF")
		env := os.Getenv("EALERT_ENV")

		if config != "" {
			fmt.Printf("trying to load diy configfile %s\n", config)
			viper.SetConfigName(config)
			if err := viper.ReadInConfig(); err != nil {
				fmt.Printf("load config file: %s error:%v\n", config, err)
				return
			}
			fmt.Printf("load config file:%v success\n", viper.ConfigFileUsed())
			return
		} else if env != "" {
			fmt.Printf("trying to load environment %s config\n", env)
			viper.SetConfigName(env)
			if err := viper.ReadInConfig(); err != nil {
				fmt.Printf("load config file:%v error:%v\n", viper.ConfigFileUsed(), err)
				return
			}
			fmt.Printf("enable env config file:%v success\n", viper.ConfigFileUsed())
		} else {
			fmt.Println("trying to load default config\n")
			viper.SetConfigName("default")
			if err := viper.ReadInConfig(); err != nil {
				fmt.Printf("load default config file error:%v\n", err)
				return
			}
			fmt.Printf("load default config success")
		}

	}
	viper.AutomaticEnv() // read in environment variables that match
}
