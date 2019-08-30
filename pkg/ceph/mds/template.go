package mds

const MdsTemp = `
kind: Deployment
apiVersion: apps/v1
metadata:
  labels:
    app: ceph
    daemon: mds
  name: ceph-mds
  namespace: {{.Namespace}}
spec:
  selector:
    matchLabels:
      app: ceph
      daemon: mds
  replicas: {{.MdsNum}}
  template:
    metadata:
      name: ceph-mds
      namespace: {{.Namespace}}
      labels:
        app: ceph
        daemon: mds
    spec:
      volumes:
        - name: ceph-configmap
          configMap:
            name: {{.CephConfName}}
        - name: ceph-conf
          emptyDir: {}
      initContainers:
      - name: ceph-init
        image: {{.CephInitImage}}
        imagePullPolicy: Always
        volumeMounts:
        - name: ceph-configmap
          mountPath: /host/ceph
        - name: ceph-conf
          mountPath: /host/etc/ceph
        command: ["/bin/sh", "-c", "sh /copy.sh"]
      containers:
        - name: ceph-mds
          image: {{.CephImage}}
          ports:
            - containerPort: 6800
          env:
            - name: CEPH_DAEMON
              value: MDS
            - name: CEPHFS_CREATE
              value: "1"
            - name: KV_TYPE
              value: k8s
            - name: CLUSTER
              value: ceph
            - name: CEPHFS_NAME
              value: "{{.CEPHFS_NAME}}"
            - name: CEPHFS_DATA_POOL
              value: "{{.CEPHFS_DATA_POOL}}"
            - name: CEPHFS_DATA_POOL_PG
              value: "128"
            - name: CEPHFS_METADATA_POOL
              value: "{{.CEPHFS_METADATA_POOL}}"
            - name: CEPHFS_METADATA_POOL_PG
              value: "128"
          volumeMounts:
            - name: ceph-conf
              mountPath: /etc/ceph
          livenessProbe:
              tcpSocket:
                port: 6800
              initialDelaySeconds: 60
              timeoutSeconds: 5
          readinessProbe:
              tcpSocket:
                port: 6800
              timeoutSeconds: 5
`
