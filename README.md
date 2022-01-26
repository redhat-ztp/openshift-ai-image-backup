# openshift-ai-image-backup
This repository provides a Dockerfile which builds the go binary in the image. The image can be locally build and pushed to **https://quay.io/repository/redhat_ztp/openshift-ai-image-backup**.

## Build:

To build the golang binary, run **./hack/build-go.sh**

## Kubernetes job

Below kubernetes job can be launched to start the backup procedure from the spoke cluster. To run this task from the hub, you need a cluster-admin role, or can create a service account and adding the proper security context to it and pass **serviceAccountName** to the below job.

```
apiVersion: batch/v1
kind: Job
metadata:
  name: container-image
automountServiceAccountToken: false
spec:
  template:
    spec:
      containers:
      - name: container-image
        image: quay.io/redhat_ztp/openshift-ai-image-backup:latest
        args: ["launchBackup", "--BackupPath", "/var/recovery"]
        securityContext:
          privileged: true
          runAsUser: 0
        tty: true
        volumeMounts:
        - name: container-backup
          mountPath: /host
      restartPolicy: Never
      volumes:
        - name: container-backup
          hostPath:
            path: /
            type: Directory
```
