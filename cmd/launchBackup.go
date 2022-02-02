/*
 * Copyright 2021 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"fmt"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

const host string = "/host"
const BackupScript string = "/usr/local/bin/cluster-backup.sh"
const etc string = "/etc/"
const usrlocal string = "/usr/local/"
const kubelet string = "/var/lib/kubelet/"

var folders []string = []string{"containers", "cluster", "etc", "usrlocal", "kubelet"}

//LaunchBackup triggers the backup procedure
// returns:			error
func LaunchBackup(BackupPath string) error {

	// check for back slash in the BackupPath
	if check := strings.Contains(BackupPath[len(BackupPath)-1:], "/"); check {
		BackupPath = BackupPath[:len(BackupPath)-1]
	}

	err := Cleanup(BackupPath)
	if err != nil {
		log.Error(fmt.Printf("Old directories couldn't be deleted, err: %s\n", err))
	}

	log.Info("Old contents have been cleaned up")

	path, err := CreateDir(BackupPath)
	if err != nil {
		log.Error(fmt.Printf("Old directories couldn't be deleted, err: %s\n", err))
	}

	//ostree pinning
	ostreeArgs := "ostree admin pin 0"
	err = ExecuteArgs(ostreeArgs, "backup-time", "ostree")
	if err != nil {
		log.Error(err)
	}

	// for now we skip the container image backup
	/*	// container image backup
		bashArgs := fmt.Sprintf("for id in $(crictl images -o json | jq -r '.images[].id'); do mkdir -p %s/$id; /usr/bin/skopeo copy --all --insecure-policy containers-storage:$id dir:%s/$id; done", path[0], path[0])
		err = ExecuteArgs(bashArgs, path[0], "containers")
		if err != nil {
			log.Warn(err)
		}
	*/
	// cluster backup
	err = ExecuteArgs(BackupScript, path[1], "etcd-cluster")
	if err != nil {
		return err
	}

	// etc back up
	etcExcludeArgs := fmt.Sprintf("cat /etc/tmpfiles.d/* | sed 's/#.*//' | awk '{print $2}' | grep '^/etc/' | sed 's#^/etc/##' > %s/etc.exclude.list; echo '.updated' >> %s/etc.exclude.list; echo 'kubernetes/manifests' >> %s/etc.exclude.list", BackupPath, BackupPath, BackupPath)
	err = ExecuteArgs(etcExcludeArgs, BackupPath, "etc-exclude-list")
	if err != nil {
		log.Error(err)
	}

	// usrlocal backup
	etcArgs := fmt.Sprintf("rsync -a %s %s", etc, path[2])
	err = ExecuteArgs(etcArgs, path[2], etc)
	if err != nil {
		log.Error(err)
	}

	usrlocalArgs := fmt.Sprintf("rsync -a %s %s", usrlocal, path[3])
	err = ExecuteArgs(usrlocalArgs, path[3], usrlocal)
	if err != nil {
		log.Error(err)
	}

	kubeletArgs := fmt.Sprintf("rsync -a %s %s", kubelet, path[4])
	err = ExecuteArgs(kubeletArgs, path[4], kubelet)
	if err != nil {
		log.Error(err)
	}

	log.Info(strings.Repeat("-", 60))
	log.Info("backup has succesfully finished ...")

	return nil

}

// Cleanup deletes all old sub diectories and files in the recovery partition
// returns: 			error
func Cleanup(path string) error {
	//change root directory to /host
	if err := syscall.Chroot(host); err != nil {
		log.Errorf("Couldn't do chroot to %s, err: %s", host, err)
		return err
	}

	log.Info(strings.Repeat("-", 60))
	log.Info("Cleaning up old contents and ostree deployment started ...")
	log.Info(strings.Repeat("-", 60))
	// Cleanup previous backups
	dir, _ := os.Open(path)
	subDir, _ := dir.Readdir(0)

	// Loop over the directory's files.
	for index := range subDir {
		fileNames := subDir[index]

		// Get name of file and its full path.
		name := fileNames.Name()
		fullPath := path + "/" + name
		log.Info("\nfullpath: ", fullPath)

		// Remove the file.
		err := os.RemoveAll(fullPath)
		if err != nil {
			log.Error(err)
			return err
		}
	}
	log.Info("Old directories deleted with contents")

	// ostree undeploy of previous deployments
	ostreeClean := "while :; do ostree admin undeploy 1 || break; done"
	err := ExecuteArgs(ostreeClean, "backup-time", "ostree-cleaning")
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

//CreateDir creates new sub-directories where backup will be stored
// returns:  slice of filepath ([]string) error
func CreateDir(path string) ([]string, error) {

	log.Info(strings.Repeat("-", 60))
	log.Info("Creating new directories and backup has been started ...")
	log.Info(strings.Repeat("-", 60))
	//create backup folders
	newPath := make([]string, len(folders))
	err := os.Chdir(path)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	for i := 0; i < len(folders); i++ {
		err := os.Mkdir(folders[i], 0700)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		newPath[i] = path + "/" + folders[i]
		log.Info("sub directory created at: ", newPath[i])
	}
	return newPath, nil
}

//ExecuteArgs execute shell commands
//returns: 			error
func ExecuteArgs(arg string, path string, resource string) error {
	if resource == "etcd-cluster" {
		_, err := exec.Command(arg, path).Output()
		if err != nil {
			log.Error(err)
			return err
		}
		return nil
	}
	_, err := exec.Command("/bin/bash", "-c", arg).Output()
	if err != nil {
		log.Error(err)
		return err
	}

	log.Info(fmt.Sprintf("%s backup has been taken at %s", resource, path))
	return nil
}

// launchBackupCmd represents the launch command
var launchBackupCmd = &cobra.Command{
	Use:   "launchBackup",
	Short: "It will trigger backup of resources in the specified path",

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

		// start launching the backup of the resource
		return LaunchBackup(BackupPath)
	},
}

func init() {

	rootCmd.AddCommand(launchBackupCmd)

	launchBackupCmd.Flags().StringP("BackupPath", "p", "", "Path where to store the backup")
	_ = launchBackupCmd.MarkFlagRequired("BackupPath")

	// bind to viper
	_ = viper.BindPFlag("BackupPath", launchBackupCmd.Flags().Lookup("BackupPath"))
}
