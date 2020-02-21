package mgr

const MgrTemp = `
kind: Deployment
apiVersion: apps/v1
metadata:
  labels:
    app: ceph
    daemon: mgr
  name: {{.MgrDpName}}
  namespace: {{.Namespace}}
spec:
  replicas: {{.MgrNum}}
  selector:
    matchLabels:
      app: ceph
      daemon: mgr
  template:
    metadata:
      name: ceph-mgr
      namespace: {{.Namespace}}
      labels:
        app: ceph
        daemon: mgr
    spec:
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
                - key: app
                  operator: In
                  values: ["ceph"]
                - key: daemon
                  operator: In
                  values: ["mgr"]
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
        - name: ceph-mgr
          image: {{.CephImage}}
          command: ["/bin/sh","-c","sh -x /etc/ceph/start_mgr.sh"]
          ports:
            - containerPort: 6800
            - containerPort: 8443
              name: dashboard
          env:
            - name: CEPH_DAEMON
              value: MGR
            - name: POD_IP
              valueFrom:
                fieldRef:
                  apiVersion: v1
                  fieldPath: status.podIP
            - name: MGR_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: DEBUG
              value: stayalive
            - name: KV_TYPE
              value: k8s
            - name: NETWORK_AUTO_DETECT
              value: "4"
            - name: CLUSTER
              value: ceph
            - name: DASHBOARD
              value: enable
          volumeMounts:
            - name: ceph-conf
              mountPath: /etc/ceph
          livenessProbe:
              tcpSocket:
                port: 6800
              initialDelaySeconds: 60
              timeoutSeconds: 5
          readinessProbe:
              tcpSocket:
                port: 6800
              timeoutSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: ceph-mgr
  namespace: {{.Namespace}}
spec:
  ports:
  - name: http
    port: 443
    protocol: TCP
    targetPort: 8443
  selector:
    app: ceph
    daemon: mgr
  type: NodePort
`
