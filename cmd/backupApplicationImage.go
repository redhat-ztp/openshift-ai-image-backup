package cmd

import (
	"fmt"
	"io"

	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"os"

	log "github.com/sirupsen/logrus"
)

func copyImage(BackupPath string) error {

	if check := strings.Contains(BackupPath[len(BackupPath)-1:], "/"); check {
		BackupPath = BackupPath[:len(BackupPath)-1]
	}

	//bash command to execute on the spoke
	bashArgs := fmt.Sprintf("for id in $(crictl images -o json | jq -r '.images[].id'); do mkdir -p %s/$id; /usr/bin/skopeo copy --all --insecure-policy containers-storage:$id dir:%s/$id; done", BackupPath, BackupPath)

	cmd := exec.Command("/bin/bash", "-c", bashArgs)
	cmd.Stdout = io.Writer(os.Stdout)

	err := cmd.Run()
	if err != nil {
		log.Error(err)
		return err
	}

	log.Info(fmt.Sprintf("Application images have been copied into %s", BackupPath))
	return nil

}

var backupApplicationImageCmd = &cobra.Command{
	Use:   "backupApplicationImage",
	Short: "It will copy the application images available under /var/lib/containers into local path specified",

	RunE: func(cmd *cobra.Command, args []string) error {

		BackupPath, _ := cmd.Flags().GetString("BackupPath")

		// validate path
		if _, err := os.Stat(BackupPath); os.IsNotExist(err) {
			// create path
			err := os.Mkdir(BackupPath, os.ModePerm)
			if err != nil {
				log.Error(err)
				return err
			}
		}

		return copyImage(BackupPath)
	},
}

func init() {
	rootCmd.AddCommand(backupApplicationImageCmd)

	backupApplicationImageCmd.Flags().StringP("BackupPath", "p", "", "Path where to copy the application images")
	backupApplicationImageCmd.MarkFlagRequired("BackupPath")

	// bind to viper
	viper.BindPFlag("BackupPath", backupApplicationImageCmd.Flags().Lookup("BackupPath"))

}
