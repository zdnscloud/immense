package mds

const MdsTemp = `
kind: Deployment
apiVersion: apps/v1
metadata:
  labels:
    app: ceph
    daemon: mds
  name: {{.MdsDpName}}
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
      tolerations:
        - key: CriticalAddonsOnly
          operator: Exists
        - key: node-role.kubernetes.io/master
          operator: Exists
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: daemon
                  operator: In
                  values: ["mds"]
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
        - name: ceph-mds
          image: {{.CephImage}}
          command: ["/bin/sh","-c","sh -x /etc/ceph/start_mds.sh"]
          ports:
            - containerPort: 6800
          env:
            - name: CEPHFS_CREATE
              value: "1"
            - name: MDS_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: FSID
              value: {{.FSID}}
            - name: MON_HOSTS
              value: "{{.MON_HOSTS}}"
            - name: MON_MEMBERS
              value: "{{.MON_MEMBERS}}"
            - name: CLUSTER
              value: ceph
            - name: CEPHFS_NAME
              value: "{{.CEPHFS_NAME}}"
            - name: CEPHFS_DATA_POOL
              value: "{{.CEPHFS_DATA_POOL}}"
            - name: CEPHFS_DATA_POOL_PG
              value: "{{.PgNum}}"
            - name: CEPHFS_METADATA_POOL
              value: "{{.CEPHFS_METADATA_POOL}}"
            - name: CEPHFS_METADATA_POOL_PG
              value: "{{.PgNum}}"
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
