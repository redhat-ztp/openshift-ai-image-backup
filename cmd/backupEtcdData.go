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

func copydata(BackupPath string) error {

	if check := strings.Contains(BackupPath[len(BackupPath)-1:], "/"); check {
		BackupPath = BackupPath[:len(BackupPath)-1]
	}

	backupScript := "/usr/local/bin/cluster-backup.sh"
	cmd := exec.Command(backupScript, BackupPath)
	cmd.Stdout = io.Writer(os.Stdout)

	err := cmd.Run()
	if err != nil {
		log.Error(err)
		return err
	}

	log.Info(fmt.Sprintf("Etcd snapshot is backed up at %s", BackupPath))
	return nil

}

var backupEtcdDataCmd = &cobra.Command{
	Use:   "backupEtcdData",
	Short: "It will back up etcd data by creating an etcd snapshot and backing up the resources for static pods",

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

		return copydata(BackupPath)
	},
}

func init() {
	rootCmd.AddCommand(backupEtcdDataCmd)

	backupEtcdDataCmd.Flags().StringP("BackupPath", "p", "", "Path where to copy the application images")
	backupEtcdDataCmd.MarkFlagRequired("BackupPath")

	// bind to viper
	viper.BindPFlag("BackupPath", backupEtcdDataCmd.Flags().Lookup("BackupPath"))

}
