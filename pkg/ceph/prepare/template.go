package prepare

const PrepareTemp = `
apiVersion: batch/v1
kind: Job
metadata:
  name: ceph-job-prepare-{{.NodeName}}
  namespace: {{.Namespace}}
spec:
  backoffLimit: 10
  template:
    metadata:
      name: ceph-job-prepare-{{.NodeName}}
    spec:
      restartPolicy: Never
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
          command: ["/bin/bash", "-c", "sh -x /etc/ceph/prepare.sh"]
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
            - name: OSD_DEVICES
              value: "{{.Devices}}"
`
