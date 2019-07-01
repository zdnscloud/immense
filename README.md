# immense

## Global Notice
* 存储节点必须是k8s集群节点
* 块设备将会强制格式化（请谨慎选择空闲磁盘）！

## Lvm
### Example
`kubectl apply -f deploy/lvm.conf`
### Notice
* 替换/删除磁盘的前提是该存储节点没有使用本地存储的Pod在运行，否则删除失败

## Ceph
### Example
`kubectl apply -f deploy/ceph.conf`

### Notice
* 集群Pod的地址段为10.42.0.0/16（固定的，因为集群暂时还没有存储该值）
* 删除磁盘时请确保集群容量>=使用容量
