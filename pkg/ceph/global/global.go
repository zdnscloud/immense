package global

const (
	NodeLabelValue          = "Ceph"
	RBAC                    = "rbac"
	MonSvc                  = "ceph-mon"
	ConfigMapName           = "ceph-conf"
	SecretName              = "csi-cephfs-secret"
	MonNum                  = 3
	MgrNum                  = 2
	MdsNum                  = 2
	StorageClassName        = "cephfs"
	CephFsName              = "myfs"
	CephFsDate              = "myfs_data"
	CephFsMetadata          = "myfs_metadata"
	CephInitImage           = "busybox:1.31.0"
	CephImage               = "ceph/daemon:latest-mimic"
	CSIAttacherImage        = "quay.io/k8scsi/csi-attacher:v1.0.1"
	CSIProvisionerImage     = "quay.io/k8scsi/csi-provisioner:v1.0.1"
	CSIDriverRegistrarImage = "quay.io/k8scsi/csi-node-driver-registrar:v1.0.2"
	CephFsCSIImage          = "quay.io/cephcsi/cephfsplugin:v1.0.0"
)
