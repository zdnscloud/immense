package config

const MonSvcTemp = `
apiVersion: v1
kind: Service
metadata:
  name: {{.SvcName}}
  namespace: {{.Namespace}}
  labels:
    app: ceph-mon
spec:
  ports:
  - port: 6789
    protocol: TCP
    targetPort: 6789
  selector:
    app: ceph-mon
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
    mon host = {{.MonHost}}
    public network = {{.Network}}
    cluster network = {{.Network}}
    osd_pool_default_size = {{.Replication}}
    osd_pool_default_min_size = 1
    osd_crush_chooseleaf_type = 0
    osd_pool_default_pg_num   = 100
    osd_pool_default_pgp_num  = 100
    rbd_default_features      = 3
    fatal_signal_handlers     = false
    mon_allow_pool_delete     = true
    mon_max_pg_per_osd        = 1000
    osd_pg_bits               = 11
    osd_pgp_bits              = 11
    osd journal size = 100
    log file = /dev/null
    osd_memory_target = 4242538496
    osd_memory_base = 2147483648
    osd_memory_cache_min = 3195011072
    cephx = true
    cephx_require_signatures = false
    cephx_cluster_require_signatures = true
    cephx_service_require_signatures = false
  ceph.client.admin.keyring: |
    [client.admin]
            key = {{.AdminKey}}
            caps mds = "allow *"
            caps mgr = "allow *"
            caps mon = "allow *"
            caps osd = "allow *"
  ceph.mon.keyring: |
    [mon.]
            key = {{.MonKey}}
            caps mon = "allow *"
    [client.admin]
            key = {{.AdminKey}}
            caps mds = "allow *"
            caps mgr = "allow *"
            caps mon = "allow *"
            caps osd = "allow *"
`

const SecretTemp = `
---
apiVersion: v1
kind: Secret
metadata:
  name: {{.CephSecretName}}
  namespace: {{.StorageNamespace}}
data:
  userID: {{.CephAdminUserEncode}}
  userKey: {{.CephAdminKeyEncode}}

  adminID: {{.CephAdminUserEncode}}
  adminKey: {{.CephAdminKeyEncode}}
`
