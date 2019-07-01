package mon

const MonTemp = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ceph-mon
  namespace: {{.Namespace}}
spec:
  replicas: {{.MonNum}}
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
          - name: MON_NAME
            valueFrom:
              fieldRef:
                fieldPath: spec.nodeName
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
