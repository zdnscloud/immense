package config

const MonSvcTemp = `
apiVersion: v1
kind: Service
metadata:
  name: {{.SvcName}}
  namespace: {{.Namespace}}
  annotations:
    service.alpha.kubernetes.io/tolerate-unready-endpoints: "true"
  labels:
    app: ceph
    daemon: mon
spec:
  ports:
  - port: {{.MonPort}}
    protocol: TCP
    targetPort: {{.MonPort}}
  selector:
    app: ceph
    daemon: mon
  clusterIP: None
`

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

    # auth
    max_open_files = 131072
    osd_pool_default_pg_num = 128
    osd_pool_default_pgp_num = 128
    osd_pool_default_size = {{.Replication}}
    osd_pool_default_min_size = 1

    mon_osd_full_ratio = .95
    mon_osd_nearfull_ratio = .85

    mon_host = {{.MonHost}}

    [mon]
    mon_osd_down_out_interval = 600
    mon_osd_min_down_reporters = 4
    mon_clock_drift_allowed = .15
    mon_clock_drift_warn_backoff = 30
    mon_osd_report_timeout = 300
    mon_data_avail_warn = 10


    [osd]
    journal_size = 100
    cluster_network = {{.Network}}
    public_network = {{.Network}}
    osd_mkfs_type = xfs
    osd_mkfs_options_xfs = -f -i size=2048
    osd_mon_heartbeat_interval = 30
    osd_max_object_name_len = 256

    #crush
    osd_pool_default_crush_rule = 0
    osd_crush_update_on_start = true

    #backend
    osd_objectstore = filestore

    #performance tuning
    filestore_merge_threshold = 40
    filestore_split_multiple = 8
    osd_op_threads = 8
    filestore_op_threads = 8
    filestore_max_sync_interval = 5
    osd_max_scrubs = 1

    #recovery tuning
    osd_recovery_max_active = 5
    osd_max_backfills = 2
    osd_recovery_op_priority = 2
    osd_client_op_priority = 63
    osd_recovery_max_chunk = 1048576
    osd_recovery_threads = 1

    #ports
    ms_bind_port_min = 6800
    ms_bind_port_max = 7100

    [client]
    rbd_cache_enabled = true
    rbd_cache_writethrough_until_flush = true
    rbd_default_features = 1

    [mds]
    mds_cache_size = 100000

    [mgr]
    client mount uid = 0
    client mount gid = 0
  ceph.client.admin.keyring: |
    [client.admin]
      key = {{.AdminKey}}
      auid = 0
      caps mds = "allow *"
      caps mon = "allow *"
      caps osd = "allow *"
      caps mgr = "allow *"
  ceph.mon.keyring: |
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
