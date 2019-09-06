package global

const (
	RBAC                    = "rbac"
	ServiceAccountName      = "ceph"
	MonDpName               = "ceph-mon"
	MdsDpName               = "ceph-mds"
	MgrDpName               = "ceph-mgr"
	MonSvc                  = "ceph-mon"
	MonPort                 = "3300"
	ConfigMapName           = "ceph-conf"
	SecretName              = "csi-cephfs-secret"
	MonNum                  = 3
	MgrNum                  = 2
	MdsNum                  = 2
	PoolDefaultSize         = 2
	CephFsName              = "myfs"
	CephFsDate              = "myfs_data"
	CephFsMetadata          = "myfs_metadata"
	CephInitImage           = "zdnscloud/ceph-init:v0.3"
	CephImage               = "ceph/daemon:latest"
	CSIProvisionerStsName   = "csi-cephfsplugin-provisioner"
	CSIPluginDsName         = "csi-cephfsplugin"
	CSIConfigmapName        = "ceph-csi-config"
	CSIAttacherImage        = "quay.io/k8scsi/csi-attacher:v1.0.1"
	CSIProvisionerImage     = "quay.io/k8scsi/csi-provisioner:v1.0.1"
	CSIDriverRegistrarImage = "quay.io/k8scsi/csi-node-driver-registrar:v1.0.2"
	CephFsCSIImage          = "quay.io/cephcsi/cephcsi:v1.1.0"
)
