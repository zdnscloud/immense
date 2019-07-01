package mds

const MdsTemp = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ceph-mds
  namespace: {{.Namespace}}
spec:
  replicas: {{.MdsNum}}
  selector:
    matchLabels:
      app: ceph-mds
  template:
    metadata:
      name: ceph-mds
      labels:
        app: ceph-mds
    spec:
      initContainers:
      - name: ceph-init
        image: {{.CephInitImage}}
        volumeMounts:
        - name: cephconf
          mountPath: /tmp/ceph
        - name: shared-data
          mountPath: /ceph
        command: ["/bin/sh", "-c", "cp /tmp/ceph/* /ceph"]
      containers:
      - name: ceph-mds
        image: {{.CephImage}}
        args:
          - "mds"
        env:
          - name: MDS_NAME
            valueFrom:
              fieldRef:
                fieldPath: spec.nodeName
          - name: CEPHFS_CREATE
            value: "1"
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
        - name: shared-data
          mountPath: /etc/ceph
      volumes:
       - name: cephconf
         configMap:
           name: ceph-conf
       - name: shared-data
         emptyDir: {}
`
