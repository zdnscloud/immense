package iscsi

const IscsiLvmdTemplate = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    app: iscsi-lvmd
  name: {{.IscsiLvmdDsName}}
  namespace: {{.StorageNamespace}}
spec:
  selector:
    matchLabels:
      app: iscsi-lvmd
  template:
    metadata:
      labels:
        app: iscsi-lvmd
    spec:
      hostNetwork: true
      nodeSelector:
        {{.LabelKey}}: {{.LabelValue}}
      initContainers:
      - name: init
        image: {{.IscsiInitImage}}
        command: ["/bin/bash", "-c", "sh -x /init.sh"]
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
          - name: TARGET_HOST
            value: "{{.TargetHost}}"
          - name: TARGET_PORT
            value: "{{.TargetPort}}"
          - name: TARGET_IQN
            value: "{{.TargetIqn}}"
          - name: VOLUME_GROUP
            value: "{{.VolumeGroup}}"
      containers:
      - name: iscsi-lvmd
        image: {{.IscsiLvmdImage}}
        command:
        - /bin/sh
        - -c
        - /lvmd
        - -listen
        - 0.0.0.0:1736
        env:
        - name: NodeName
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
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
          path: /lib/modules
      - hostPath:
          path: /run/udev
          type: ""
        name: run-udev`

const IscsiCSITemplate = `
{{- if eq .RBACConfig "rbac"}}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: csi-iscsiplugin-provisioner
  namespace: {{.StorageNamespace}}
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: csi-iscsiplugin-provisioner
  namespace: {{.StorageNamespace}}
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["list", "watch"]
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
  name: csi-iscsiplugin-provisioner-role
  namespace: {{.StorageNamespace}}
subjects:
  - kind: ServiceAccount
    name: csi-iscsiplugin-provisioner
    namespace: {{.StorageNamespace}}
roleRef:
  kind: ClusterRole
  name: csi-iscsiplugin-provisioner
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: csi-iscsiplugin
  namespace: {{.StorageNamespace}}
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: csi-iscsiplugin
  namespace: {{.StorageNamespace}}
rules:
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "update", "watch"]
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["list", "watch"]
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
  name: csi-iscsiplugin
  namespace: {{.StorageNamespace}}
subjects:
  - kind: ServiceAccount
    name: csi-iscsiplugin
    namespace: {{.StorageNamespace}}
roleRef:
  kind: ClusterRole
  name: csi-iscsiplugin
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
      app: csi-iscsiplugin
  template:
    metadata:
      labels:
        app: csi-iscsiplugin
    spec:
      nodeSelector: 
        {{.LabelKey}}: {{.LabelValue}}
      serviceAccount: csi-iscsiplugin
      hostNetwork: true
      containers:
        - name: csi-iscsiplugin-driver-registrar
          image: {{.CSIDriverRegistrarImage}}
          args:
            - "--v=5"
            - "--csi-address=$(ADDRESS)"
            - "--kubelet-registration-path=/var/lib/kubelet/plugins/csi-iscsi/csi.sock"
          lifecycle:
            preStop:
              exec:
                command: ["/bin/sh", "-c", "rm -rf /registration/csi-iscsiplugin-reg.sock /csi/"]
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
            - "--labelKey={{.LabelKey}}"
            - "--labelValue={{.LabelValue}}"
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
            path: /var/lib/kubelet/plugins/csi-iscsi/
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
  name: csi-iscsiplugin-provisioner
  namespace: {{.StorageNamespace}}
  labels:
    app: csi-iscsiplugin-provisioner
spec:
  selector:
    app: csi-iscsiplugin-provisioner
  ports:
    - name: dummy
      port: 12345
---
kind: StatefulSet
apiVersion: apps/v1
metadata:
  name: {{.IscsiCSIStsName}}
  namespace: {{.StorageNamespace}}
spec:
  serviceName: csi-iscsiplugin-provisioner
  replicas: 1
  selector:
    matchLabels:
      app: csi-iscsiplugin-provisioner
  template:
    metadata:
      labels:
        app: csi-iscsiplugin-provisioner
    spec:
      serviceAccount: csi-iscsiplugin-provisioner
      nodeSelector: 
        {{.LabelKey}}: {{.LabelValue}}
      hostNetwork: true
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
            - "--labelKey={{.LabelKey}}"
            - "--labelValue={{.LabelValue}}"
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
