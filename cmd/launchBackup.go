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

func launchBackup(BackupPath string) error {

	// check for back slash in the BackupPath
	if check := strings.Contains(BackupPath[len(BackupPath)-1:], "/"); check {
		BackupPath = BackupPath[:len(BackupPath)-1]
	}

	//change root directory to /host
	if err := syscall.Chroot(host); err != nil {
		log.Println("Couldn't do chroot to %s", host)
		return nil
	}

	log.Info(fmt.Println(strings.Repeat("-", 45)))
	log.Info(fmt.Println("Cleaning up old contents and ostree deployment started ..."))
	log.Info(fmt.Println(strings.Repeat("-", 45)))

	// Cleanup previous backups
	dir, _ := os.Open(BackupPath)
	subDir, _ := dir.Readdir(0)

	// Loop over the directory's files.
	for index := range subDir {
		fileNames := subDir[index]

		// Get name of file and its full path.
		name := fileNames.Name()
		fullPath := BackupPath + "/" + name
		log.Info(fmt.Printf("fullpath: %s", fullPath))

		if subDir[index].Name() != "extras.tgz" {
			// Remove the file.
			os.RemoveAll(fullPath)
		}

	}
	log.Info(fmt.Println("Old directories deleted with contents"))

	// ostree undeploy of previous deployments
	ostreeClean := "while :; do ostree admin undeploy 1 || break; done"
	_, err := exec.Command("/bin/bash", "-c", ostreeClean).Output()
	if err != nil {
		log.Error(err)
		//	return err
	}
	log.Info(fmt.Println("Ostree pin undeployed"))

	//create backup folders
	//position				0			1		   2		3			4
	folders := []string{"containers", "cluster", "etc", "usrlocal", "kubelet"}
	path := make([]string, len(folders))

	os.Chdir(BackupPath)

	for i := 0; i < len(folders); i++ {
		err := os.Mkdir(folders[i], os.ModePerm)
		if err != nil {
			log.Error(err)
			//		return err
		}
		path[i] = BackupPath + "/" + folders[i]
		log.Info(fmt.Printf("file created: %s", path[i]))
	}

	log.Info(fmt.Println(strings.Repeat("-", 45)))
	log.Info(fmt.Println("Old contents have been cleaned up"))
	log.Info(fmt.Println(strings.Repeat("-", 45)))
	// ------------- end cleanup ------------

	// ------------- start backup ------------
	log.Info(fmt.Println("New contents will be backed up with new ostree deployment..."))
	log.Info(fmt.Println(strings.Repeat("-", 45)))
	//ostree pinning
	ostreeArgs := "ostree admin pin 0"
	_, err = exec.Command("/bin/bash", "-c", ostreeArgs).Output()
	if err != nil {
		log.Error(err)
		//	return err
	}
	log.Info(fmt.Println("ostree backup has been created"))

	// container image backup
	bashArgs := fmt.Sprintf("for id in $(crictl images -o json | jq -r '.images[].id'); do mkdir -p %s/$id; /usr/bin/skopeo copy --all --insecure-policy containers-storage:$id dir:%s/$id; done", path[0], path[0])
	_, err = exec.Command("/bin/bash", "-c", bashArgs).Output()
	if err != nil {
		log.Warn(err)
	}
	log.Info(fmt.Sprintf("Application images have been copied into %s", path[0]))

	// cluster backup
	_, err = exec.Command(BackupScript, path[1]).Output()
	if err != nil {
		log.Error(err)
		return err
	}
	log.Info(fmt.Sprintf("Etcd snapshot is backed up at %s", BackupPath))

	// etc back up
	etcExcludeArgs := fmt.Sprintf("cat /etc/tmpfiles.d/* | sed 's/#.*//' | awk '{print $2}' | grep '^/etc/' | sed 's#^/etc/##' > %s/etc.exclude.list; echo '.updated' >> %s/etc.exclude.list; echo 'kubernetes/manifests' >> %s/etc.exclude.list", BackupPath, BackupPath, BackupPath)
	_, err = exec.Command("/bin/bash", "-c", etcExcludeArgs).Output()
	if err != nil {
		log.Error(err)
		return err
	}
	log.Info(fmt.Sprintf("etc.exclude.list has been updated under %s", BackupPath))

	// usrlocal backup
	etcArgs := fmt.Sprintf("rsync -a %s %s", etc, path[2])
	_, err = exec.Command("/bin/bash", "-c", etcArgs).Output()
	if err != nil {
		log.Error(err)
		return err
	}
	log.Info(fmt.Sprintf("/etc dir backup has been taken into %s", path[2]))

	// usrlocal backup
	usrlocalArgs := []string{"-a", usrlocal, path[3]}
	_, err = exec.Command("rsync", usrlocalArgs...).Output()
	if err != nil {
		log.Error(err)
		return err
	}
	log.Info(fmt.Sprintf("/usr/local dir backup has been taken into %s/%s", BackupPath, path[3]))

	kubeletArgs := []string{"-a", kubelet, path[4]}
	_, err = exec.Command("rsync", kubeletArgs...).Output()
	if err != nil {
		log.Error(err)
		return err
	}
	log.Info(fmt.Sprintf("/var/lib/kubelet dir backup has been taken into %s/%s", BackupPath, path[4]))

	log.Info(fmt.Println(strings.Repeat("-", 45)))
	log.Info(fmt.Println("backup has succesfully finished ..."))

	return nil

}

// launchBackupCmd represents the launch command
var launchBackupCmd = &cobra.Command{
	Use:   "launchBackup",
	Short: "It will read the rootfs url and will copy into given folder ",

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
		return launchBackup(BackupPath)
	},
}

func init() {

	rootCmd.AddCommand(launchBackupCmd)

	launchBackupCmd.Flags().StringP("BackupPath", "p", "", "Path where to store the backup")
	launchBackupCmd.MarkFlagRequired("BackupPath")

	// bind to viper
	viper.BindPFlag("BackupPath", launchBackupCmd.Flags().Lookup("BackupPath"))
}
