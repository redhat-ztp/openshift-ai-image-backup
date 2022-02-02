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
      hostNetwork: true
      restartPolicy: Never
      volumes:
        - name: container-backup
          hostPath:
            path: /
            type: Directory
```

## Launch the backup from hub with manage cluster action

To launch this job as managed cluster action from the hub, one need to create a namespace, service account, clusterrolebinding and the job using managed cluster action:

For example:

**namespace.yaml**

```
apiVersion: action.open-cluster-management.io/v1beta1
kind: ManagedClusterAction
metadata:
  name: mca-namespace
  namespace: snonode-virt02
spec:
  actionType: Create
  kube:
    resource: namespace
    template:
      apiVersion: v1
      kind: Namespace
      metadata:
        name: backupresource
```


**serviceAccount.yaml**

```
apiVersion: action.open-cluster-management.io/v1beta1
kind: ManagedClusterAction
metadata:
  name: mca-serviceaccount
  namespace: snonode-virt02
spec:
  actionType: Create
  kube:
    resource: serviceaccount
    template:
      apiVersion: v1
      kind: ServiceAccount
      metadata:
        name: backupresource
        namespace: backupresource

```

**clusterrolebinding.yaml**
```
apiVersion: action.open-cluster-management.io/v1beta1
kind: ManagedClusterAction
metadata:
  name: mca-rolebinding
  namespace: snonode-virt02
spec:
  actionType: Create
  kube:
    resource: clusterrolebinding
    template:
      apiVersion: rbac.authorization.k8s.io/v1
      kind: ClusterRoleBinding
      metadata:
        name: backupResource
      roleRef:
        apiGroup: rbac.authorization.k8s.io
        kind: ClusterRole
        name: cluster-admin
      subjects:
        - kind: ServiceAccount
          name: backupresource
          namespace: backupresource

```


**k8sJob.yaml**
```
apiVersion: action.open-cluster-management.io/v1beta1
kind: ManagedClusterAction
metadata:
  name: mca-ob
  namespace: snonode-virt02
spec:
  actionType: Create
  kube:
    namespace: backupresource
    resource: job
    template:
      apiVersion: batch/v1
      kind: Job
      metadata:
        name: backupresource
      spec:
        template:
          spec:
            containers:
              - 
                args:
                  - launchBackup
                  - "--BackupPath"
                  - /var/recovery
                image: quay.io/redhat_ztp/openshift-ai-image-backup:latest
                name: container-image
                securityContext:
                  privileged: true
                  runAsUser: 0
                tty: true
                volumeMounts:
                  - 
                    mountPath: /host
                    name: backup
            hostNetwork: true
            restartPolicy: Never
            serviceAccountName: backupresource
            volumes:
              - 
                hostPath:
                  path: /
                  type: Directory
                name: backup

```
