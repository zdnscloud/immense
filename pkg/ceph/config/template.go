package config

const MonSvcTemp = `
apiVersion: v1
kind: Service
metadata:
  labels:
    mon_svc: {{.MonSvc}}
    mon_id: {{.MonID}}
  name: {{.MonSvc}}-{{.MonID}}
  namespace: {{.Namespace}}
spec:
  ports:
  - name: msgr1
    port: {{.MonPortV1}}
    protocol: TCP
    targetPort: {{.MonPortV1}}
  - name: msgr2
    port: {{.MonPortV2}}
    protocol: TCP
    targetPort: {{.MonPortV2}}
  selector:
    mon_svc: {{.MonSvc}}
    mon_id: {{.MonID}}
  type: ClusterIP`

const ConfigMapTemp = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{.CephConfName}}
  namespace: {{.Namespace}}
data:
  ceph.conf: |
    [global]
    fsid = {{.FSID}}
    cephx = true
    cephx_require_signatures = false
    cephx_cluster_require_signatures = true
    cephx_service_require_signatures = false
    osd_memory_target = 4242538496
    osd_memory_base = 2147483648
    osd_memory_cache_min = 3195011072
    mon_max_pg_per_osd = 1000

    # auth
    max_open_files = 131072
    osd_pool_default_size = {{.Replication}}
    osd_pool_default_min_size = 1

    mon_osd_full_ratio = .95
    mon_osd_nearfull_ratio = .85

    mon initial members       = {{.MonMembers}}
    mon host                  = {{.MonEp}}

    [mon]
    mon_clock_drift_allowed = .15
    mon_clock_drift_warn_backoff = 30
    mon_data_avail_warn = 10

    #performance tuning
    filestore_merge_threshold = 40
    filestore_split_multiple = 8
    filestore_op_threads = 8
    filestore_max_sync_interval = 5

    #recovery tuning
    osd_recovery_max_active = 5
    #ports
    ms_bind_port_min = 6800
    ms_bind_port_max = 7100

    [mds]
    mds_cache_size = 100000

  ceph.client.admin.keyring: |
    [client.admin]
      key = {{.AdminKey}}
      auid = 0
      caps mds = "allow *"
      caps mon = "allow *"
      caps osd = "allow *"
      caps mgr = "allow *"
  keyring: |
    [client.admin]
      key = {{.AdminKey}}
      caps mds = "allow *"
      caps mon = "allow *"
      caps osd = "allow *"
      caps mgr = "allow *"
    [mon.]
      key = {{.MonKey}}
      caps mon = "allow *"

`

const SecretTemp = `
---
apiVersion: v1
kind: Secret
metadata:
  name: {{.CephSecretName}}
  namespace: {{.StorageNamespace}}
stringData:
  userID: {{.CephAdminUser}}
  userKey: {{.CephAdminKey}}

  adminID: {{.CephAdminUser}}
  adminKey: {{.CephAdminKey}}
`
const ServiceAccountTemp = `
{{- if eq .RBACConfig "rbac"}}
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{.CephSAName}}
  namespace: {{.StorageNamespace}}
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: {{.CephSAName}}
  namespace: {{.StorageNamespace}}
rules:
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get", "list"]

---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: {{.CephSAName}}
  namespace: {{.StorageNamespace}}
subjects:
  - kind: ServiceAccount
    name: {{.CephSAName}}
    namespace: {{.StorageNamespace}}
roleRef:
  kind: ClusterRole
  name: {{.CephSAName}}
  apiGroup: rbac.authorization.k8s.io
{{- end}}
`
