package zap

const OsdZapTemp = `
apiVersion: batch/v1
kind: Job
metadata:
  name: ceph-job-zap-{{.NodeName}}-{{.OsdID}}
  namespace: {{.Namespace}}
spec:
  backoffLimit: 10
  template:
    metadata:
      name: ceph-job-zap-{{.NodeName}}-{{.OsdID}}
    spec:
      restartPolicy: Never
      nodeName: "{{.NodeName}}"
      containers:
      - name: ceph-zap
        image: {{.CephImage}}
        lifecycle:
          postStart:
            exec:
              command: ["/bin/sh","-c","for i in $(dmsetup ls|grep ceph|grep osd|awk '{print $1}');do dmsetup remove -f $i;done"]
        command: ["/bin/sh","-c","/opt/ceph-container/bin/entrypoint.sh zap_device;dd if=/dev/zero of=$(OSD_DEVICE) bs=1M count=1024"]
        securityContext:
          privileged: true
          capabilities:
            add: ["SYS_ADMIN"]
          allowPrivilegeEscalation: true
        env:
          - name: OSD_DEVICE
            value: "/dev/{{.OsdID}}"
        volumeMounts:
        - name: dev
          mountPath: /dev
      volumes:
       - name: dev
         hostPath:
           path: /dev
`
