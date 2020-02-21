package global

const (
	StorageType             = "cephfs"
	StorageDriverName       = "cephfs.csi.ceph.com"
	RBAC                    = "rbac"
	ServiceAccountName      = "ceph"
	MonDpName               = "ceph-mon"
	MdsDpName               = "ceph-mds"
	MgrDpName               = "ceph-mgr"
	MonSvc                  = "ceph-mon"
	MonPortV1               = "6789"
	MonPortV2               = "3300"
	ConfigMapName           = "ceph-conf"
	SecretName              = "csi-cephfs-secret"
	MgrNum                  = 1
	MdsNum                  = 2
	PoolDefaultSize         = 2
	TargetPgPerOsd          = 100
	PgNumDefault            = 128
	CephFsName              = "myfs"
	CephFsDate              = "myfs_data"
	CephFsMetadata          = "myfs_metadata"
	CephInitImage           = "zdnscloud/ceph-init:v0.7"
	CephImage               = "ceph/ceph:v14.2.7-20200206"
	CSIProvisionerStsName   = "csi-cephfsplugin-provisioner"
	CSIPluginDsName         = "csi-cephfsplugin"
	CSIConfigmapName        = "ceph-csi-config"
	CSIAttacherImage        = "quay.io/k8scsi/csi-attacher:v2.1.0"
	CSIResizerImage         = "quay.io/k8scsi/csi-resizer:v0.4.0"
	CSIProvisionerImage     = "quay.io/k8scsi/csi-provisioner:v1.4.0"
	CSIDriverRegistrarImage = "quay.io/k8scsi/csi-node-driver-registrar:v1.2.0"
	CephFsCSIImage          = "quay.io/cephcsi/cephcsi:v2.0.0"
)

var MonMembers = []string{"a", "b", "c"}
