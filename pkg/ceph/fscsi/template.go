package fscsi

const FSconfigmapTemp = `
apiVersion: v1
kind: ConfigMap
data:
  config.json: |-
    [
      {
        "clusterID": "{{.CephClusterID}}",
        "monitors": [{{.CephMons}}]
      }
    ]
metadata:
  name: {{.CSIConfigmapName}}
  namespace: {{.StorageNamespace}}
`

const FScsiTemp = `
{{- if eq .RBACConfig "rbac"}}
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cephfs-csi-provisioner
  namespace: {{.StorageNamespace}}
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cephfs-external-provisioner-runner
  namespace: {{.StorageNamespace}}
rules:
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["list", "watch", "create", "update", "patch"]
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "create", "delete"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["csi.storage.k8s.io"]
    resources: ["csinodeinfos"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["volumeattachments"]
    verbs: ["get", "list", "watch", "update"]

---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cephfs-csi-provisioner-role
  namespace: {{.StorageNamespace}}
subjects:
  - kind: ServiceAccount
    name: cephfs-csi-provisioner
    namespace: {{.StorageNamespace}}
roleRef:
  kind: ClusterRole
  name: cephfs-external-provisioner-runner
  apiGroup: rbac.authorization.k8s.io

---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cephfs-external-provisioner-cfg
  namespace: {{.StorageNamespace}}
rules:
  - apiGroups: [""]
    resources: ["endpoints"]
    verbs: ["get", "watch", "list", "delete", "update", "create"]
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get", "list", "create", "delete"]

---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cephfs-csi-provisioner-role-cfg
  namespace: {{.StorageNamespace}}
subjects:
  - kind: ServiceAccount
    name: cephfs-csi-provisioner
    namespace: {{.StorageNamespace}}
roleRef:
  kind: Role
  name: cephfs-external-provisioner-cfg
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cephfs-csi-nodeplugin
  namespace: {{.StorageNamespace}}
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cephfs-csi-nodeplugin
  namespace: {{.StorageNamespace}}
rules:
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "update"]
  - apiGroups: [""]
    resources: ["namespaces"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["volumeattachments"]
    verbs: ["get", "list", "watch", "update"]

---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cephfs-csi-nodeplugin
  namespace: {{.StorageNamespace}}
subjects:
  - kind: ServiceAccount
    name: cephfs-csi-nodeplugin
    namespace: {{.StorageNamespace}}
roleRef:
  kind: ClusterRole
  name: cephfs-csi-nodeplugin
  apiGroup: rbac.authorization.k8s.io
{{- end}}
---
kind: Service
apiVersion: v1
metadata:
  name: csi-cephfsplugin-provisioner
  namespace: {{.StorageNamespace}}
  labels:
    app: csi-cephfsplugin-provisioner
spec:
  selector:
    app: csi-cephfsplugin-provisioner
  ports:
    - name: dummy
      port: 12345

---
kind: StatefulSet
apiVersion: apps/v1
metadata:
  name: {{.CSIProvisionerStsName}}
  namespace: {{.StorageNamespace}}
spec:
  selector:
    matchLabels:
      app: csi-cephfsplugin-provisioner
  serviceName: "csi-cephfsplugin-provisioner"
  replicas: 1
  template:
    metadata:
      labels:
        app: csi-cephfsplugin-provisioner
    spec:
      serviceAccount: cephfs-csi-provisioner
      containers:
        - name: csi-provisioner
          image: {{.StorageCephProvisionerImage}}
          args:
            - "--csi-address=$(ADDRESS)"
            - "--v=5"
          env:
            - name: ADDRESS
              value: unix:///csi/csi-provisioner.sock
          imagePullPolicy: "IfNotPresent"
          volumeMounts:
            - name: socket-dir
              mountPath: /csi
        - name: csi-cephfsplugin-attacher
          image: {{.StorageCephAttacherImage}}
          args:
            - "--v=5"
            - "--csi-address=$(ADDRESS)"
          env:
            - name: ADDRESS
              value: /csi/csi-provisioner.sock
          imagePullPolicy: "IfNotPresent"
          volumeMounts:
            - name: socket-dir
              mountPath: /csi
        - name: csi-cephfsplugin
          securityContext:
            privileged: true
            capabilities:
              add: ["SYS_ADMIN"]
          # for stable functionality replace v1.1.0 with latest release version
          image: {{.StorageCephFsCSIImage}}
          args:
            - "--nodeid=$(NODE_ID)"
            - "--type=cephfs"
            - "--endpoint=$(CSI_ENDPOINT)"
            - "--v=5"
            - "--drivername=cephfs.csi.ceph.com"
            - "--metadatastorage=k8s_configmap"
          env:
            - name: NODE_ID
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            - name: POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: CSI_ENDPOINT
              value: unix:///csi/csi-provisioner.sock
          imagePullPolicy: "IfNotPresent"
          volumeMounts:
            - name: socket-dir
              mountPath: /csi
            - name: host-sys
              mountPath: /sys
            - name: lib-modules
              mountPath: /lib/modules
              readOnly: true
            - name: host-dev
              mountPath: /dev
            - name: ceph-csi-config
              mountPath: /etc/ceph-csi-config/
      volumes:
        - name: socket-dir
          hostPath:
            path: /var/lib/kubelet/plugins/cephfs.csi.ceph.com
            type: DirectoryOrCreate
        - name: host-sys
          hostPath:
            path: /sys
        - name: lib-modules
          hostPath:
            path: /lib/modules
        - name: host-dev
          hostPath:
            path: /dev
        - name: ceph-csi-config
          configMap:
            name: {{.CSIConfigmapName}}
---
kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: {{.CSIPluginDsName}}
  namespace: {{.StorageNamespace}}
spec:
  selector:
    matchLabels:
      app: csi-cephfsplugin
  template:
    metadata:
      labels:
        app: csi-cephfsplugin
    spec:
      serviceAccount: cephfs-csi-nodeplugin
      hostNetwork: true
      dnsPolicy: ClusterFirstWithHostNet
      containers:
        - name: driver-registrar
          image: {{.StorageCephDriverRegistrarImage}}
          args:
            - "--v=5"
            - "--csi-address=/csi/csi.sock"
            - "--kubelet-registration-path=/var/lib/kubelet/plugins/cephfs.csi.ceph.com/csi.sock"
          lifecycle:
            preStop:
              exec:
                command: [
                  "/bin/sh", "-c",
                  "rm -rf /registration/csi-cephfsplugin \
                  /registration/csi-cephfsplugin-reg.sock"
                ]
          env:
            - name: KUBE_NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
          volumeMounts:
            - name: plugin-dir
              mountPath: /csi
            - name: registration-dir
              mountPath: /registration
        - name: csi-cephfsplugin
          securityContext:
            privileged: true
            capabilities:
              add: ["SYS_ADMIN"]
            allowPrivilegeEscalation: true
          image: {{.StorageCephFsCSIImage}}
          args:
            - "--nodeid=$(NODE_ID)"
            - "--type=cephfs"
            - "--endpoint=$(CSI_ENDPOINT)"
            - "--v=5"
            - "--drivername=cephfs.csi.ceph.com"
            - "--metadatastorage=k8s_configmap"
            - "--mountcachedir=/mount-cache-dir"
          env:
            - name: NODE_ID
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            - name: POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: CSI_ENDPOINT
              value: unix:///csi/csi.sock
          imagePullPolicy: "IfNotPresent"
          volumeMounts:
            - name: mount-cache-dir
              mountPath: /mount-cache-dir
            - name: plugin-dir
              mountPath: /csi
            - name: csi-plugins-dir
              mountPath: /var/lib/kubelet/plugins/kubernetes.io/csi
              mountPropagation: "Bidirectional"
            - name: pods-mount-dir
              mountPath: /var/lib/kubelet/pods
              mountPropagation: "Bidirectional"
            - name: host-sys
              mountPath: /sys
            - name: lib-modules
              mountPath: /lib/modules
              readOnly: true
            - name: host-dev
              mountPath: /dev
            - name: ceph-csi-config
              mountPath: /etc/ceph-csi-config/
      volumes:
        - name: mount-cache-dir
          emptyDir: {}
        - name: plugin-dir
          hostPath:
            path: /var/lib/kubelet/plugins/cephfs.csi.ceph.com/
            type: DirectoryOrCreate
        - name: csi-plugins-dir
          hostPath:
            path: /var/lib/kubelet/plugins/kubernetes.io/csi
            type: DirectoryOrCreate
        - name: registration-dir
          hostPath:
            path: /var/lib/kubelet/plugins_registry/
            type: Directory
        - name: pods-mount-dir
          hostPath:
            path: /var/lib/kubelet/pods
            type: Directory
        - name: host-sys
          hostPath:
            path: /sys
        - name: lib-modules
          hostPath:
            path: /lib/modules
        - name: host-dev
          hostPath:
            path: /dev
        - name: ceph-csi-config
          configMap:
            name: {{.CSIConfigmapName}}
`
const StorageClassTemp = `
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: {{.StorageClassName}}
provisioner: cephfs.csi.ceph.com
parameters:
  clusterID: {{.CephClusterID}}
  fsName: {{.CephFsName}}
  pool: {{.CephFsPool}}
  csi.storage.k8s.io/provisioner-secret-name: {{.CephSecretName}}
  csi.storage.k8s.io/provisioner-secret-namespace: {{.StorageNamespace}}
  csi.storage.k8s.io/node-stage-secret-name: {{.CephSecretName}}
  csi.storage.k8s.io/node-stage-secret-namespace: {{.StorageNamespace}}

reclaimPolicy: Delete
`
