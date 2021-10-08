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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
	"net/url"
	"os"
	"os/exec"
)

// sync image with skopeo
func syncImage(ReleaseImageURL string, BackupPath string, AuthFilePath string) error {
	// first remove all content in that folder
	os.RemoveAll(BackupPath)

	skopeoArgs := []string{"sync", "--src", "docker", "--dest", "dir", ReleaseImageURL, BackupPath}

	if len(AuthFilePath) > 0 {
		// add the authfile flag
		skopeoArgs = append(skopeoArgs, []string{"--authfile", AuthFilePath}...)
	}

	_, err := exec.Command("/usr/bin/skopeo", skopeoArgs...).Output()
    if err != nil {
        log.Error(err)
		return err
    }
	log.Info(fmt.Sprintf("Sync completed for %s into %s", ReleaseImageURL, BackupPath))
	return nil
}

var backupReleaseImageCmd = &cobra.Command{
	Use:   "backupReleaseImage",
	Short: "It will sync the release image specified into local path specified",

	RunE: func(cmd *cobra.Command, args []string) error {
		ReleaseImageURL, _ := cmd.Flags().GetString("ReleaseImageURL")
		BackupPath, _ := cmd.Flags().GetString("BackupPath")

		// validate url
		_, err := url.ParseRequestURI("docker://"+ReleaseImageURL)
		if err != nil {
			log.Error(err)
			return err	
		}
		
		// validate path
		if _, err := os.Stat(BackupPath); os.IsNotExist(err) {
			// create path
			err := os.Mkdir(BackupPath, os.ModePerm)
			if err != nil {
				log.Error(err)
				return err
			}
		}

		// start syncing image
		AuthFilePath := ""
		if cmd.Flags().Changed("AuthFilePath") {
			AuthFilePath, _ = cmd.Flags().GetString("AuthFilePath")
		}
		return syncImage(ReleaseImageURL, BackupPath, AuthFilePath)
	},
}

func init() {
	rootCmd.AddCommand(backupReleaseImageCmd)

	backupReleaseImageCmd.Flags().StringP("ReleaseImageURL", "r", "", "URL for the release image to sync")
	backupReleaseImageCmd.Flags().StringP("BackupPath", "p", "", "Path where to sync the release image")
	backupReleaseImageCmd.Flags().StringP("AuthFilePath", "a", "", "Path to the registry authentication file")
	backupReleaseImageCmd.MarkFlagRequired("ReleaseImageURL")
	backupReleaseImageCmd.MarkFlagRequired("BackupPath")

	// bind to viper
	viper.BindPFlag("ReleaseImageURL", backupReleaseImageCmd.Flags().Lookup("ReleaseImageURL"))
	viper.BindPFlag("BackupPath", backupReleaseImageCmd.Flags().Lookup("BackupPath"))
	viper.BindPFlag("AuthFilePath", backupReleaseImageCmd.Flags().Lookup("AuthFilePath"))
}
