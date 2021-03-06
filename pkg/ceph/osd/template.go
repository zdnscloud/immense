package osd

const OsdTemp = `
kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: ceph-osd-{{.NodeName}}-{{.OsdID}}
  namespace: {{.Namespace}}
  labels:
    app: ceph
    daemon: osd-{{.NodeName}}-{{.OsdID}}
spec:
  selector:
    matchLabels:
      app: ceph
      daemon: osd-{{.NodeName}}-{{.OsdID}}
  template:
    metadata:
      name: ceph-osd-{{.NodeName}}-{{.OsdID}}
      labels:
        app: ceph
        daemon: osd-{{.NodeName}}-{{.OsdID}}
    spec:
      hostPID: true
      nodeSelector:
        kubernetes.io/hostname: "{{.NodeName}}"
      volumes:
        - name: devices
          hostPath:
            path: /dev
        - name: ceph-configmap
          configMap:
            name: {{.CephConfName}}
        - name: ceph-conf
          emptyDir: {}
        - hostPath:
            path: /run/udev
            type: ""
          name: run-udev
        - hostPath:
            path: /var/lib/ceph
            type: ""
          name: ceph-data
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
        - name: osd-pod
          image: {{.CephImage}}
          command: ["/bin/bash", "-c", "sh -x /etc/ceph/start_osd.sh"]
          imagePullPolicy: "IfNotPresent"
          volumeMounts:
            - name: devices
              mountPath: /dev
            - name: ceph-conf
              mountPath: /etc/ceph
            - mountPath: /run/udev
              name: run-udev
          securityContext:
            privileged: true
            readOnlyRootFilesystem: false
            runAsUser: 0
          env:
            - name: OSD_TYPE
              value: "disk"
            - name: OSD_DEVICE
              value: "/dev/{{.OsdID}}"
            - name: KV_TYPE
              value: k8s
            - name: CLUSTER
              value: ceph
            - name: CEPH_GET_ADMIN_KEY
              value: "1"
            - name: OSD_BLUESTORE
              value: "1"
            - name: FSID
              value: {{.FSID}}
            - name: MON_MEMBERS
              value: {{.Mon_Members}}
            - name: MON_ENDPOINT
              value: {{.Mon_Endpoint}}
            - name: ADDR
              valueFrom:
                fieldRef:
                  fieldPath: status.podIP
            - name: OSD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
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
