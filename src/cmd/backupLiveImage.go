/*
Copyright © 2021 Yolanda Robla <yroblamo@redhat.com>

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

func copyResource(RootFSURL string, BackupPath string, filename string) error {
	// first remove all content in that folder
	os.RemoveAll(BackupPath)

	_, err := grab.Get(fmt.Sprintf("%s/%s", BackupPath, filename), RootFSURL)
	if err != nil {
		log.Error(err)
		return err
	} 
	log.Info(fmt.Sprintf("Download completed for %s into %s/%s", RootFSURL, BackupPath, filename))
	return nil
}

// backupLiveImageCmd represents the backupLiveImage command
var backupLiveImageCmd = &cobra.Command{
	Use:   "backupLiveImage",
	Short: "It will read the rootfs url and will copy into given folder ",

	RunE: func(cmd *cobra.Command, args []string) error {
		RootFSURL, _ := cmd.Flags().GetString("RootFSURL")
		BackupPath, _ := cmd.Flags().GetString("BackupPath")

		// validate url
		result, err := url.ParseRequestURI(RootFSURL)
		if err != nil {
			log.Error(err)
			return err	
		}

		path := result.Path
		filename := path[strings.LastIndex(path, "/")+1:]
		
		// validate path
		if _, err := os.Stat(BackupPath); os.IsNotExist(err) {
			// create path
			err := os.Mkdir(BackupPath, os.ModePerm)
			if err != nil {
				log.Error(err)
				return err
			}
		}

		// start copying the resource
		return copyResource(RootFSURL, BackupPath, filename)
	},
}

func init() {
	
	rootCmd.AddCommand(backupLiveImageCmd)

	backupLiveImageCmd.Flags().StringP("RootFSURL", "u", "", "URL from where to download the live rootfs img")
	backupLiveImageCmd.Flags().StringP("BackupPath", "p", "", "Path where to store the rootfs url")
	backupLiveImageCmd.MarkFlagRequired("RootFSURL")
	backupLiveImageCmd.MarkFlagRequired("BackupPath")

	// bind to viper
	viper.BindPFlag("RootFSURL", backupLiveImageCmd.Flags().Lookup("RootFSURL"))
	viper.BindPFlag("BackupPath", backupLiveImageCmd.Flags().Lookup("BackupPath"))
}
