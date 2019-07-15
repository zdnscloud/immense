package mgr

const MgrTemp = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ceph-mgr
  namespace: {{.Namespace}}
spec:
  replicas: {{.MgrNum}}
  selector:
    matchLabels:
      app: ceph-mgr
  template:
    metadata:
      name: ceph-mgr
      labels:
        app: ceph-mgr
    spec:
      initContainers:
      - name: ceph-init-conf
        image: {{.CephInitImage}}
        volumeMounts:
        - name: cephconf
          mountPath: /tmp/ceph
        - name: shared-data
          mountPath: /ceph
        command: ["/bin/sh", "-c", "cp /tmp/ceph/* /ceph"]
      - name: ceph-init-health
        image: {{.CephImage}}
        command: ["/bin/sh", "-c", "until ceph health --connect-timeout 10|grep HEALTH_OK; do echo waiting for ceph cluster to health; sleep 2; done;"]
        volumeMounts:
        - name: shared-data
          mountPath: /etc/ceph
      containers:
      - name: ceph-mgr
        image: {{.CephImage}}
        args:
          - "mgr"
        env:
          - name: MGR_NAME
            valueFrom:
              fieldRef:
                fieldPath: spec.nodeName
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
