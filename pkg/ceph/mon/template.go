package mon

const MonTemp = `
kind: StatefulSet
apiVersion: apps/v1beta2
metadata:
  name: ceph-mon
  namespace: {{.Namespace}}
spec:
  replicas: {{.MonNum}}
  serviceName: ceph-mon
  selector:
    matchLabels:
      app: ceph-mon
  template:
    metadata:
      name: ceph-mon
      labels:
        app: ceph-mon
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
      - name: ceph-mon
        image: {{.CephImage}}
        args:
          - "mon"
        env:
          - name: MON_IP
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: status.podIP
          - name: CEPH_PUBLIC_NETWORK
            value: {{.Network}}
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
