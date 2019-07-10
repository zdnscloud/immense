# immense

## Global Notice
* 存储节点必须是k8s集群节点
* 块设备将会强制格式化（请谨慎选择空闲磁盘）！
* 目前storageType只支持lvm、ceph两种，且每一种只能有一个cluster
* lvm类型存储的storageclass为lvm;ceph类型存的storgeclass为cephfs

## Lvm
### Example
`kubectl apply -f deploy/lvm.conf`
### Notice
* 替换/删除磁盘的时如果有Pod在使用该节点的本地存储，会导致这个Pod异常。

## Ceph
### Example
`kubectl apply -f deploy/ceph.conf`

### Notice
* 配置使用了集群Pod的默认地址段为10.42.0.0/16（固定的，因为集群暂时还没有存储该值）
* 删除磁盘时请确保集群容量>=使用容量
* 更换磁盘的过程较长（因为ceph集群需要同步数据）
* 如果磁盘少于2块，副本数将为1。默认为2
* 建议不要同时删除多个磁盘，否则可能会导致部分数据无法恢复。删除一块之后需要等待集群状态为HEALTH_OK方可继续删除
