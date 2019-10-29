package mon

const MonTemp = `
kind: Deployment
apiVersion: apps/v1
metadata:
  labels:
    mon_svc: {{.MonSvc}}
    mon_id: {{.ID}}
  name: ceph-mon-{{.ID}}
  namespace: {{.Namespace}}
spec:
  replicas: 1
  selector:
    matchLabels:
      mon_svc: {{.MonSvc}}
      mon_id: {{.ID}}
  template:
    metadata:
      name: ceph-mon-{{.MonID}}
      namespace: {{.Namespace}}
      labels:
        mon_svc: {{.MonSvc}}
        mon_id: {{.ID}}
    spec:
      serviceAccount: {{.ServiceAccountName}}
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
                - key: mon_svc
                  operator: In
                  values: ["{{.MonSvc}}"]
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
        - name: ceph-mon
          image: {{.CephImage}}
          command: ["/bin/sh", "-c", "sh -x /etc/ceph/start_mon.sh"]
          ports:
            - containerPort: {{.MonPortV1}}
            - containerPort: {{.MonPortV2}}
          env:
            - name: FSID
              value: {{.FSID}}
            - name: MON_HOSTS
              value: "{{.MON_HOSTS}}"
            - name: MON_MEMBERS
              value: "{{.MON_MEMBERS}}"
            - name: ID
              value: {{.ID}}
            - name: MON_SVC_ADDR
              value: {{.MonSvcAddr}}
            - name: MON_IP
              valueFrom:
                fieldRef:
                  fieldPath: status.podIP
          volumeMounts:
            - name: ceph-conf
              mountPath: /etc/ceph
          livenessProbe:
              tcpSocket:
                port: {{.MonPortV1}}
              initialDelaySeconds: 60
              timeoutSeconds: 5
          readinessProbe:
              tcpSocket:
                port: {{.MonPortV1}}
              timeoutSeconds: 5
`
