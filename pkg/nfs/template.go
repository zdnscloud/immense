package nfs

const NfsCSITemplate = `
{{- if eq .RBACConfig "rbac"}}
apiVersion: v1
kind: ServiceAccount
metadata:
  name: nfs-client-provisioner
  namespace: {{.StorageNamespace}}
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: nfs-client-provisioner-runner
  namespace: {{.StorageNamespace}}
rules:
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "create", "delete"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["create", "update", "patch"]
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: run-nfs-client-provisioner
  namespace: {{.StorageNamespace}}
subjects:
  - kind: ServiceAccount
    name: nfs-client-provisioner
    namespace: {{.StorageNamespace}}
roleRef:
  kind: ClusterRole
  name: nfs-client-provisioner-runner
  apiGroup: rbac.authorization.k8s.io
---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: leader-locking-nfs-client-provisioner
    # replace with namespace where provisioner is deployed
  namespace: {{.StorageNamespace}}
rules:
  - apiGroups: [""]
    resources: ["endpoints"]
    verbs: ["get", "list", "watch", "create", "update", "patch"]
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: leader-locking-nfs-client-provisioner
  namespace: {{.StorageNamespace}}
subjects:
  - kind: ServiceAccount
    name: nfs-client-provisioner
    namespace: {{.StorageNamespace}}
roleRef:
  kind: Role
  name: leader-locking-nfs-client-provisioner
  apiGroup: rbac.authorization.k8s.io
{{- end}}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.NfsCSIDpName}}
  labels:
    app: nfs-client-provisioner
  namespace: {{.StorageNamespace}}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nfs-client-provisioner
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: nfs-client-provisioner
  template:
    metadata:
      labels:
        app: nfs-client-provisioner
    spec:
      serviceAccountName: nfs-client-provisioner
      containers:
        - name: nfs-client-provisioner
          image: {{.NFSProvisionerImage}}
          volumeMounts:
            - name: nfs-client-root
              mountPath: /persistentvolumes
          env:
            - name: PROVISIONER_NAME
              value: {{.ProvisionerName}}
            - name: NFS_SERVER
              value: {{.NfsServer}}
            - name: NFS_PATH
              value: {{.NfsPath}}
      volumes:
        - name: nfs-client-root
          nfs:
            server: {{.NfsServer}}
            path: {{.NfsPath}}`

const StorageClassTemp = `
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: {{.StorageClassName}}
provisioner: {{.ProvisionerName}}
parameters:
  archiveOnDelete: "false"`
