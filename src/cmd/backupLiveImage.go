/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (

	"fmt"
	"github.com/cavaliercoder/grab"	
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
	"net/url"
	"os"
	"strings"
)

var (
	RootFSUrl string
	BackupPath string
)

// copyyResource will download RootFSUrl and copy to BackupPath
func copyResource(filename string) error {
	_, err := grab.Get(fmt.Sprintf("%s/%s", BackupPath, filename), RootFSUrl)
	if err != nil {
		log.Error(err)
		return err
	} 
	log.Info(fmt.Sprintf("Download completed for %s into %s/%s", RootFSUrl, BackupPath, filename))
	return nil
}

// backupLiveImageCmd represents the backupLiveImage command
var backupLiveImageCmd = &cobra.Command{
	Use:   "backupLiveImage",
	Short: "It will read the rootfs url and will copy into given folder ",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		// validate url
		result, err := url.ParseRequestURI(RootFSUrl)
		if err != nil {
			log.Error(err)
			return err	
		}

		path := result.Path
		filename := path[strings.LastIndex(path, "/")+1:]
		
		// validate path
		if _, err := os.Stat(BackupPath); os.IsNotExist(err) {
			// path/to/whatever does not exist
			log.Error(err)
			return err
		}

		// start copying the resource
		return copyResource(filename)
	},
}

func init() {
	rootCmd.AddCommand(backupLiveImageCmd)

	backupLiveImageCmd.Flags().StringVarP(&RootFSUrl, "rootFSURL", "u", "", "URL from where to download the live rootfs img")
	backupLiveImageCmd.Flags().StringVarP(&BackupPath, "BackupPath", "p", "", "Path where to store the rootfs url")
	backupLiveImageCmd.MarkFlagRequired("rootFSURL")
	backupLiveImageCmd.MarkFlagRequired("BackupPath")

	// bind to viper
	viper.BindPFlag("rootFSURL", backupLiveImageCmd.Flags().Lookup("rootFSURL"))
	viper.BindPFlag("BackupPath", backupLiveImageCmd.Flags().Lookup("backupPath"))
}
