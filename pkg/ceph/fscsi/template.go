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
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cephfs-csi-nodeplugin
  namespace: {{.StorageNamespace}}
---
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
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["list", "watch", "create", "update", "patch"]
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "create", "delete", "patch"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["volumeattachments"]
    verbs: ["get", "list", "watch", "update", "patch"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims/status"]
    verbs: ["update", "patch"]
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cephfs-csi-provisioner-role
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
    resources: ["configmaps"]
    verbs: ["get", "list", "create", "delete"]
  - apiGroups: ["coordination.k8s.io"]
    resources: ["leases"]
    verbs: ["get", "watch", "list", "delete", "update", "create"]
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
{{- end}}
---
kind: Service
apiVersion: v1
metadata:
  name: csi-cephfsplugin-provisioner
  namespace: {{.StorageNamespace}}
  labels:
    app: csi-metrics
spec:
  selector:
    app: csi-cephfsplugin-provisioner
  ports:
    - name: http-metrics
      port: 8080
      protocol: TCP
      targetPort: 8681
    - name: grpc-metrics
      port: 8090
      protocol: TCP
      targetPort: 8091
---
kind: StatefulSet
apiVersion: apps/v1
metadata:
  name: {{.CSIProvisionerStsName}}
  namespace: {{.StorageNamespace}}
spec:
  serviceName: csi-cephfsplugin-provisioner
  selector:
    matchLabels:
      app: csi-cephfsplugin-provisioner
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
            - "--timeout=150s"
            - "--enable-leader-election=true"
            - "--leader-election-type=leases"
            - "--retry-interval-start=500ms"
          env:
            - name: ADDRESS
              value: unix:///csi/csi-provisioner.sock
          imagePullPolicy: "IfNotPresent"
          volumeMounts:
            - name: socket-dir
              mountPath: /csi
        - name: csi-resizer
          image: {{.StorageCephResizerImage}}
          args:
            - "--csi-address=$(ADDRESS)"
            - "--v=5"
            - "--csiTimeout=150s"
            - "--leader-election"
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
            - "--leader-election=true"
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
          image: {{.StorageCephFsCSIImage}}
          args:
            - "--nodeid=$(NODE_ID)"
            - "--type=cephfs"
            - "--controllerserver=true"
            - "--endpoint=$(CSI_ENDPOINT)"
            - "--v=5"
            - "--drivername={{.StorageDriverName}}"
            - "--metadatastorage=k8s_configmap"
            - "--pidlimit=-1"
            - "--metricsport=8091"
            - "--metricspath=/metrics"
            - "--enablegrpcmetrics=false"
          env:
            - name: POD_IP
              valueFrom:
                fieldRef:
                  fieldPath: status.podIP
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
            - name: keys-tmp-dir
              mountPath: /tmp/csi/keys
        - name: liveness-prometheus
          image: {{.StorageCephFsCSIImage}}
          args:
            - "--type=liveness"
            - "--endpoint=$(CSI_ENDPOINT)"
            - "--metricsport=8681"
            - "--metricspath=/metrics"
            - "--polltime=60s"
            - "--timeout=3s"
          env:
            - name: CSI_ENDPOINT
              value: unix:///csi/csi-provisioner.sock
            - name: POD_IP
              valueFrom:
                fieldRef:
                  fieldPath: status.podIP
          volumeMounts:
            - name: socket-dir
              mountPath: /csi
          imagePullPolicy: "IfNotPresent"
      volumes:
        - name: socket-dir
          emptyDir: {
            medium: "Memory"
          }
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
            name: ceph-csi-config
        - name: keys-tmp-dir
          emptyDir: {
            medium: "Memory"
          }
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
          securityContext:
            privileged: true
          image: {{.StorageCephDriverRegistrarImage}}
          args:
            - "--v=5"
            - "--csi-address=/csi/csi.sock"
            - "--kubelet-registration-path=/var/lib/kubelet/plugins/{{.StorageDriverName}}/csi.sock"
          lifecycle:
            preStop:
              exec:
                command: [
                  "/bin/sh", "-c",
                  "rm -rf /registration/{{.StorageDriverName}}-reg.sock /csi/"
                ]
          env:
            - name: KUBE_NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
          volumeMounts:
            - name: socket-dir
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
            - "--nodeserver=true"
            - "--endpoint=$(CSI_ENDPOINT)"
            - "--v=5"
            - "--drivername={{.StorageDriverName}}"
            - "--metadatastorage=k8s_configmap"
            - "--metricsport=8090"
            - "--metricspath=/metrics"
            - "--enablegrpcmetrics=false"
          env:
            - name: POD_IP
              valueFrom:
                fieldRef:
                  fieldPath: status.podIP
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
            - name: socket-dir
              mountPath: /csi
            - name: mountpoint-dir
              mountPath: /var/lib/kubelet/pods
              mountPropagation: Bidirectional
            - name: plugin-dir
              mountPath: /var/lib/kubelet/plugins
              mountPropagation: "Bidirectional"
            - name: host-sys
              mountPath: /sys
            - name: lib-modules
              mountPath: /lib/modules
              readOnly: true
            - name: host-dev
              mountPath: /dev
            - name: host-mount
              mountPath: /run/mount
            - name: ceph-csi-config
              mountPath: /etc/ceph-csi-config/
            - name: keys-tmp-dir
              mountPath: /tmp/csi/keys
        - name: liveness-prometheus
          securityContext:
            privileged: true
          image: {{.StorageCephFsCSIImage}}
          args:
            - "--type=liveness"
            - "--endpoint=$(CSI_ENDPOINT)"
            - "--metricsport=8681"
            - "--metricspath=/metrics"
            - "--polltime=60s"
            - "--timeout=3s"
          env:
            - name: CSI_ENDPOINT
              value: unix:///csi/csi.sock
            - name: POD_IP
              valueFrom:
                fieldRef:
                  fieldPath: status.podIP
          volumeMounts:
            - name: socket-dir
              mountPath: /csi
          imagePullPolicy: "IfNotPresent"
      volumes:
        - name: socket-dir
          hostPath:
            path: /var/lib/kubelet/plugins/{{.StorageDriverName}}/
            type: DirectoryOrCreate
        - name: registration-dir
          hostPath:
            path: /var/lib/kubelet/plugins_registry/
            type: Directory
        - name: mountpoint-dir
          hostPath:
            path: /var/lib/kubelet/pods
            type: DirectoryOrCreate
        - name: plugin-dir
          hostPath:
            path: /var/lib/kubelet/plugins
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
        - name: host-mount
          hostPath:
            path: /run/mount
        - name: ceph-csi-config
          configMap:
            name: ceph-csi-config
        - name: keys-tmp-dir
          emptyDir: {
            medium: "Memory"
          }
---
apiVersion: v1
kind: Service
metadata:
  name: csi-metrics-cephfsplugin
  namespace: {{.StorageNamespace}}
  labels:
    app: csi-metrics
spec:
  ports:
    - name: http-metrics
      port: 8080
      protocol: TCP
      targetPort: 8681
    - name: grpc-metrics
      port: 8090
      protocol: TCP
      targetPort: 8091
  selector:
    app: csi-cephfsplugin
`
const StorageClassTemp = `
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: {{.StorageClassName}}
provisioner: {{.StorageDriverName}}
parameters:
  clusterID: {{.CephClusterID}}
  fsName: {{.CephFsName}}
  csi.storage.k8s.io/provisioner-secret-name: {{.CephSecretName}}
  csi.storage.k8s.io/provisioner-secret-namespace: {{.StorageNamespace}}
  csi.storage.k8s.io/node-stage-secret-name: {{.CephSecretName}}
  csi.storage.k8s.io/node-stage-secret-namespace: {{.StorageNamespace}}
  csi.storage.k8s.io/controller-expand-secret-name: {{.CephSecretName}}
  csi.storage.k8s.io/controller-expand-secret-namespace: {{.StorageNamespace}}
  accessMode: ReadWriteMany
allowVolumeExpansion: true
reclaimPolicy: Delete
`
