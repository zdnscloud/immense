package iscsi

const IscsiInitTemplate = `
apiVersion: apps/v1
kind: DaemonSet 
metadata:
  labels:
    app: iscsi-init-{{.Instance}}
  name: {{.IscsiInitDsName}}
  namespace: {{.StorageNamespace}}
spec:
  selector:
    matchLabels:
      app: iscsi-init-{{.Instance}}
  template:
    metadata:
      labels:
        app: iscsi-init-{{.Instance}}
    spec:
      hostNetwork: true
      nodeSelector:
        {{.IscsiInstanceLabelKey}}: "{{.IscsiInstanceLabelValue}}"
      containers:
      - name: init
        image: {{.IscsiInitImage}}
        command: ["/bin/bash", "-c", "sh -x /init.sh"]
        imagePullPolicy: "IfNotPresent"
        volumeMounts:
          - name: host-dev
            mountPath: /dev
          - mountPath: /run/udev
            name: run-udev
{{- if .CHAPConfig}}
          - mountPath: /root/secret
            name: iscsipwd
            readOnly: true
{{- end}}
        securityContext:
          privileged: true
          capabilities:
            add: ["SYS_ADMIN"]
          allowPrivilegeEscalation: true
          readOnlyRootFilesystem: false
          runAsUser: 0
        env:
          - name: TARGET_HOST
            value: "{{.TargetHost}}"
          - name: TARGET_PORT
            value: "{{.TargetPort}}"
          - name: TARGET_IQN
            value: "{{.TargetIqn}}"
          - name: VOLUME_GROUP
            value: "{{.VolumeGroup}}"
      volumes:
{{- if .CHAPConfig}}
      - name: iscsipwd
        secret:
          secretName: {{.IscsiInstanceSecret}}
{{- end}}
      - hostPath:
          path: /dev
          type: ""
        name: host-dev
      - hostPath:
          path: /run/udev
          type: ""
        name: run-udev`

const IscsiLvmdTemplate = `
kind: Service
apiVersion: v1
metadata:
  name: iscsi-lvmd-{{.Instance}}
  namespace: {{.StorageNamespace}}
  labels:
    app: iscsi-lvmd-{{.Instance}}
spec:
  selector:
    app: iscsi-lvmd-{{.Instance}}
  ports:
    - name: lvmd
      port: 1736
---
apiVersion: apps/v1
kind: DaemonSet 
metadata:
  labels:
    app: iscsi-lvmd-{{.Instance}}
  name: {{.IscsiLvmdDsName}}
  namespace: {{.StorageNamespace}}
spec:
  selector:
    matchLabels:
      app: iscsi-lvmd-{{.Instance}}
  template:
    metadata:
      labels:
        app: iscsi-lvmd-{{.Instance}}
    spec:
      nodeSelector:
        {{.IscsiInstanceLabelKey}}: "{{.IscsiInstanceLabelValue}}"
      containers:
      - name: iscsi-lvmd
        image: {{.IscsiLvmdImage}}
        command:
        - /bin/sh
        - -c
        - /lvmd
        - -listen
        - 0.0.0.0:1736
        securityContext:
          allowPrivilegeEscalation: true
          capabilities:
            add:
            - SYS_ADMIN
          privileged: true
        volumeMounts:
        - mountPath: /dev
          name: host-dev
        - mountPath: /lib/modules
          name: lib-modules
          readOnly: true
        livenessProbe:
          failureThreshold: 3
          initialDelaySeconds: 60
          periodSeconds: 10
          successThreshold: 1
          tcpSocket:
            port: 1736
          timeoutSeconds: 5
        readinessProbe:
          failureThreshold: 3
          periodSeconds: 10
          successThreshold: 1
          tcpSocket:
            port: 1736
          timeoutSeconds: 5
      volumes:
      - hostPath:
          path: /dev
          type: ""
        name: host-dev
      - name: lib-modules
        hostPath:
          path: /lib/modules`

const IscsiCSITemplate = `
{{- if eq .RBACConfig "rbac"}}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: csi-iscsiplugin-provisioner-{{.Instance}}
  namespace: {{.StorageNamespace}}
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: csi-iscsiplugin-provisioner-{{.Instance}}
  namespace: {{.StorageNamespace}}
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["list", "watch", "get"]
  - apiGroups: ["apps"]
    resources: ["daemonsets"]
    verbs: ["list", "watch", "get"]
  - apiGroups: ["apps"]
    resources: ["statefulsets"]
    verbs: ["list", "watch"]
  - apiGroups: [""]
    resources: ["namespaces"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "create", "delete", "patch"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims/status"]
    verbs: ["update", "patch"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses", "volumeattachments", "csinodes"]
    verbs: ["get", "list", "watch", "update", "patch"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["list", "watch", "create", "update", "patch"]
  - apiGroups: ["coordination.k8s.io"]
    resources: ["leases"]
    verbs: ["get", "watch", "list", "delete", "update", "create"]
  - apiGroups: ["extensions"]
    resourceNames:
    - privileged 
    resources: ["podsecuritypolicies"]
    verbs:
    - use
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: csi-iscsiplugin-provisioner-role-{{.Instance}}
  namespace: {{.StorageNamespace}}
subjects:
  - kind: ServiceAccount
    name: csi-iscsiplugin-provisioner-{{.Instance}}
    namespace: {{.StorageNamespace}}
roleRef:
  kind: ClusterRole
  name: csi-iscsiplugin-provisioner-{{.Instance}}
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: csi-iscsiplugin-{{.Instance}}
  namespace: {{.StorageNamespace}}
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: csi-iscsiplugin-{{.Instance}}
  namespace: {{.StorageNamespace}}
rules:
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "update", "watch"]
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["list", "watch", "get"]
  - apiGroups: ["apps"]
    resources: ["daemonsets"]
    verbs: ["list", "watch", "get"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims"]
    verbs: ["list", "watch"]
  - apiGroups: ["apps"]
    resources: ["statefulsets"]
    verbs: ["list", "watch"]
  - apiGroups: [""]
    resources: ["namespaces"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["volumeattachments"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["coordination.k8s.io"]
    resources: ["leases"]
    verbs: ["get", "watch", "list", "delete", "update", "create"]
  - apiGroups: ["extensions"]
    resourceNames:
    - privileged 
    resources: ["podsecuritypolicies"]
    verbs:
    - use
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: csi-iscsiplugin-{{.Instance}}
  namespace: {{.StorageNamespace}}
subjects:
  - kind: ServiceAccount
    name: csi-iscsiplugin-{{.Instance}}
    namespace: {{.StorageNamespace}}
roleRef:
  kind: ClusterRole
  name: csi-iscsiplugin-{{.Instance}}
  apiGroup: rbac.authorization.k8s.io  
{{- end}}
---
kind: DaemonSet
apiVersion: apps/v1
metadata:
  namespace: {{.StorageNamespace}}
  name: {{.IscsiCSIDsName}}
spec:
  selector:
    matchLabels:
      app: csi-iscsiplugin-{{.Instance}}
  template:
    metadata:
      labels:
        app: csi-iscsiplugin-{{.Instance}}
    spec:
      nodeSelector: 
        {{.IscsiInstanceLabelKey}}: "{{.IscsiInstanceLabelValue}}"
      serviceAccount: csi-iscsiplugin-{{.Instance}}
      containers:
        - name: csi-iscsiplugin-driver-registrar
          image: {{.CSIDriverRegistrarImage}}
          args:
            - "--v=5"
            - "--csi-address=$(ADDRESS)"
            - "--kubelet-registration-path=/var/lib/kubelet/plugins/{{.IscsiDriverName}}/csi.sock"
          lifecycle:
            preStop:
              exec:
                command: ["/bin/sh", "-c", "rm -rf /registration/{{.IscsiDriverName}}-reg.sock /csi/"]
          env:
            - name: ADDRESS
              value: /csi/csi.sock
          volumeMounts:
            - name: plugin-dir
              mountPath: /csi
            - name: registration-dir
              mountPath: /registration
        - name: csi-iscsiplugin
          securityContext:
            privileged: true
            capabilities:
              add: ["SYS_ADMIN"]
            allowPrivilegeEscalation: true
          image: {{.IscsiPluginImage}}
          args :
            - "--nodeid=$(NODE_ID)"
            - "--endpoint=$(CSI_ENDPOINT)"
            - "--vgname=$(VG_NAME)"
            - "--drivername={{.IscsiDriverName}}"
            - "--labelKey={{.IscsiInstanceLabelKey}}"
            - "--labelValue={{.IscsiInstanceLabelValue}}"
            - "--lvmdDsName={{.LvmdDsName}}"
          env:
            - name: VG_NAME
              value: "{{.VolumeGroup}}"
            - name: NODE_ID
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            - name: CSI_ENDPOINT
              value: unix://csi/csi.sock
          imagePullPolicy: "IfNotPresent"
          volumeMounts:
            - name: plugin-dir
              mountPath: /csi
            - name: pods-mount-dir
              mountPath: /var/lib/kubelet/pods
              mountPropagation: "Bidirectional"
            - mountPath: /dev
              name: host-dev
            - mountPath: /sys
              name: host-sys
            - mountPath: /lib/modules
              name: lib-modules
              readOnly: true
      volumes:
        - name: registration-dir
          hostPath:
            path: /var/lib/kubelet/plugins_registry/
            type: DirectoryOrCreate
        - name: pods-mount-dir
          hostPath:
            path: /var/lib/kubelet/pods
            type: Directory
        - name: plugin-dir
          hostPath:
            path: /var/lib/kubelet/plugins/{{.IscsiDriverName}}/
            type: DirectoryOrCreate
        - name: host-dev
          hostPath:
            path: /dev
        - name: host-sys
          hostPath:
            path: /sys
        - name: lib-modules
          hostPath:
            path: /lib/modules
---
kind: Service
apiVersion: v1
metadata:
  name: csi-iscsiplugin-provisioner-{{.Instance}}
  namespace: {{.StorageNamespace}}
  labels:
    app: csi-iscsiplugin-provisioner-{{.Instance}}
spec:
  selector:
    app: csi-iscsiplugin-provisioner-{{.Instance}}
  ports:
    - name: dummy
      port: 12345
---
kind: Deployment
apiVersion: apps/v1
metadata:
  name: {{.IscsiCSIDpName}}
  namespace: {{.StorageNamespace}}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: csi-iscsiplugin-provisioner-{{.Instance}}
  template:
    metadata:
      labels:
        app: csi-iscsiplugin-provisioner-{{.Instance}}
    spec:
      serviceAccount: csi-iscsiplugin-provisioner-{{.Instance}}
      containers:
        - name: csi-resizer
          image: {{.CSIResizerImage}}
          args:
            - "--v=5"
            - "--csi-address=$(ADDRESS)"
            - "--leader-election"
          env:
            - name: ADDRESS
              value: /csi/csi.sock
          imagePullPolicy: "IfNotPresent"
          volumeMounts:
            - name: socket-dir
              mountPath: /csi
        - name: csi-iscsiplugin-attacher
          image: {{.CSIAttacherImage}}
          args:
            - "--v=5"
            - "--csi-address=$(ADDRESS)"
            - "--leader-election=true"
          env:
            - name: ADDRESS
              value: /csi/csi.sock
          imagePullPolicy: "IfNotPresent"
          volumeMounts:
            - name: socket-dir
              mountPath: /csi
        - name: csi-iscsiplugin-provisioner
          image: {{.CSIProvisionerImage}}
          args:
            - "--csi-address=$(ADDRESS)"
            - "--v=50"
            - "--logtostderr"
            - "--feature-gates=Topology=true"
            - "--enable-leader-election=true"
            - "--leader-election-type=leases"
            - "--retry-interval-start=500ms"
          env:
            - name: ADDRESS
              value: /csi/csi.sock
          imagePullPolicy: "IfNotPresent"
          volumeMounts:
            - name: socket-dir
              mountPath: /csi
        - name: csi-iscsiplugin
          securityContext:
            privileged: true
            capabilities:
              add: ["SYS_ADMIN"]
            allowPrivilegeEscalation: true
          image: {{.IscsiPluginImage}}
          args :
            - "--nodeid=$(NODE_ID)"
            - "--endpoint=$(CSI_ENDPOINT)"
            - "--vgname=$(VG_NAME)"
            - "--drivername={{.IscsiDriverName}}"
            - "--labelKey={{.IscsiInstanceLabelKey}}"
            - "--labelValue={{.IscsiInstanceLabelValue}}"
            - "--lvmdDsName={{.LvmdDsName}}"
          env:
            - name: VG_NAME
              value: "{{.VolumeGroup}}"
            - name: NODE_ID
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            - name: CSI_ENDPOINT
              value: unix://csi/csi.sock
          imagePullPolicy: "IfNotPresent"
          volumeMounts:
            - name: socket-dir
              mountPath: /csi
            - mountPath: /dev
              name: host-dev
            - mountPath: /sys
              name: host-sys
            - mountPath: /lib/modules
              name: lib-modules
              readOnly: true
      volumes:
        - name: socket-dir
          emptyDir: {}
        - name: host-dev
          hostPath:
            path: /dev
        - name: host-sys
          hostPath:
            path: /sys
        - name: lib-modules
          hostPath:
            path: /lib/modules`

const StorageClassTemp = `
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: {{.StorageClassName}}
provisioner: {{.IscsiDriverName}}
reclaimPolicy: Delete
parameters:
  accessMode: ReadWriteOnce
#allowVolumeExpansion: true`

const IscsiStopJobTemplate = `
apiVersion: batch/v1
kind: Job
metadata:
  name: {{.IscsiStopJobName}}
  namespace: {{.StorageNamespace}}
spec:
  backoffLimit: 10
  template:
    metadata:
      name: iscsi-job-stop-{{.Host}}
    spec:
      hostNetwork: true
      restartPolicy: Never
      nodeSelector:
        kubernetes.io/hostname: "{{.Host}}"
      containers:
      - name: stop
        image: {{.IscsiInitImage}}
        command: ["/bin/bash", "-c", "iscsiadm -m node -T ${TARGET_IQN} -u"]
        imagePullPolicy: "IfNotPresent"
        volumeMounts:
          - name: host-dev
            mountPath: /dev
          - mountPath: /run/udev
            name: run-udev
        securityContext:
          privileged: true
          readOnlyRootFilesystem: false
          runAsUser: 0
        env:
          - name: TARGET_IQN
            value: "{{.TargetIqn}}"
      volumes:
      - hostPath:
          path: /dev
          type: ""
        name: host-dev
      - hostPath:
          path: /run/udev
          type: ""
        name: run-udev`
