package mon

const MonTemp = `
kind: Deployment
apiVersion: apps/v1
metadata:
  labels:
    app: ceph
    daemon: mon
  name: ceph-mon
  namespace: {{.Namespace}}
spec:
  replicas: {{.MonNum}}
  selector:
    matchLabels:
      app: ceph
      daemon: mon
  template:
    metadata:
      name: ceph-mon
      namespace: {{.Namespace}}
      labels:
        app: ceph
        daemon: mon
    spec:
      serviceAccount: {{.ServiceAccountName}}
      tolerations:
        - key: CriticalAddonsOnly
          operator: Exists
        - key: node-role.kubernetes.io/master
          operator: Exists
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: app
                operator: In
                values: ["ceph"]
              - key: daemon
                operator: In
                values: ["mon"]
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
          lifecycle:
            preStop:
                exec:
                  # remove the mon on Pod stop.
                  command:
                    - "/remove-mon.sh"
          ports:
            - containerPort: {{.MonPort}}
          env:
            - name: CEPH_DAEMON
              value: MON
            - name: KV_TYPE
              value: k8s
            - name: NETWORK_AUTO_DETECT
              value: "4"
            - name: CLUSTER
              value: ceph
          volumeMounts:
            - name: ceph-conf
              mountPath: /etc/ceph
          livenessProbe:
              tcpSocket:
                port: {{.MonPort}}
              initialDelaySeconds: 60
              timeoutSeconds: 5
          readinessProbe:
              tcpSocket:
                port: {{.MonPort}}
              timeoutSeconds: 5
`
