package osd

const OsdTemp = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: ceph-osd-{{.NodeName}}-{{.OsdID}}
  namespace: {{.Namespace}}
spec:
  selector:
    matchLabels:
      app: ceph-osd-{{.NodeName}}-{{.OsdID}}
  template:
    metadata:
      name: ceph-osd-{{.NodeName}}-{{.OsdID}}
      labels:
        app: ceph-osd-{{.NodeName}}-{{.OsdID}}
    spec:
      nodeSelector:
        kubernetes.io/hostname: {{.NodeName}}
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
      - name: ceph-osd
        image: {{.CephImage}}
        securityContext:
          privileged: true
          capabilities:
            add: ["SYS_ADMIN"]
          allowPrivilegeEscalation: true
        command: ["/bin/sh", "-c", "ceph auth get client.bootstrap-osd -o /var/lib/ceph/bootstrap-osd/ceph.keyring;/opt/ceph-container/bin/entrypoint.sh osd"]
        env:
          - name: OSD_FORCE_ZAP
            value: "1"
          - name: OSD_TYPE
            value: "disk"
          - name: OSD_DEVICE
            value: "/dev/{{.OsdID}}"
          - name: OSD_NAME
            valueFrom:
              fieldRef:
                fieldPath: spec.nodeName
        volumeMounts:
        - name: dev
          mountPath: /dev
        - name: shared-data
          mountPath: /etc/ceph
      volumes:
       - name: cephconf
         configMap:
           name: ceph-conf
       - name: dev
         hostPath:
           path: /dev
       - name: shared-data
         emptyDir: {}
`
