package mgr

const MgrTemp = `
kind: Deployment
apiVersion: apps/v1
metadata:
  labels:
    app: ceph
    daemon: mgr
  name: {{.MgrDpName}}
  namespace: {{.Namespace}}
spec:
  replicas: {{.MgrNum}}
  selector:
    matchLabels:
      app: ceph
      daemon: mgr
  template:
    metadata:
      name: ceph-mgr
      namespace: {{.Namespace}}
      labels:
        app: ceph
        daemon: mgr
    spec:
      tolerations:
        - key: CriticalAddonsOnly
          operator: Exists
        - key: node-role.kubernetes.io/master
          operator: Exists
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: app
                operator: In
                values: ["ceph"]
              - key: daemon
                operator: In
                values: ["mgr"]
            topologyKey: kubernetes.io/hostname
      volumes:
        - name: ceph-configmap
          configMap:
            name: {{.CephConfName}}
        - name: ceph-conf
          emptyDir: {}
      initContainers:
      - name: ceph-init
        image: {{.CephInitImage}}
        imagePullPolicy: "IfNotPresent"
        volumeMounts:
        - name: ceph-configmap
          mountPath: /host/ceph
        - name: ceph-conf
          mountPath: /host/etc/ceph
        command: ["/bin/sh", "-c", "sh /copy.sh"]
      containers:
        - name: ceph-mgr
          image: {{.CephImage}}
          ports:
            - containerPort: 6800
            - containerPort: 7000
              name: dashboard
          env:
            - name: CEPH_DAEMON
              value: MGR
            - name: DEBUG
              value: stayalive
            - name: KV_TYPE
              value: k8s
            - name: NETWORK_AUTO_DETECT
              value: "4"
            - name: CLUSTER
              value: ceph
          volumeMounts:
            - name: ceph-conf
              mountPath: /etc/ceph
`
