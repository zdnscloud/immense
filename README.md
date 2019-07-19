# immense

## Global Notice
* 存储节点必须是k8s集群节点
* 存储节点上没有分区、没有文件系统、没有挂载点的块设备才会在创建存储节点时自动识别加入存储
* 目前storageType只支持lvm、ceph两种，且每一种存储类型只能有一个cluster
* lvm类型存储的storageclass为lvm;ceph类型存储的storgeclass为cephfs
* 存储状态更新和监控检查频率为60秒/次

## Lvm
### Example
`kubectl apply -f deploy/lvm.conf`
### Notice
* 删除存储节点的时如果有Pod在使用该节点的本地存储，会导致磁盘清理失败

## Ceph
### Example
`kubectl apply -f deploy/ceph.conf`

### Notice
* 配置使用了集群Pod的默认地址段为10.42.0.0/16（固定的，因为集群暂时还没有存储该值）
* 删除存储节点时请确保集群容量>=使用容量
* 删除存储节点会导致ceph集群长时间处于warning状态（因为ceph集群需要同步数据）。如果该节点有多块块设备，有可能会导致数据无法恢复
* 如果总的可用块设备少于2块，副本数将为1。默认为2
* 建议不要同时删除多个节点，否则可能会导致部分数据无法恢复。删除之后需要等待集群状态为HEALTH_OK方可继续删除

# TODO
- ceph当前版本为13，需要升级到14（nautilus）。且使用的ceph-disk官方已经不推荐使用
- ceph存储节点的删除是即时的，存储数据丢失的风险
- 存储节点块设备的自动发现及扩容
