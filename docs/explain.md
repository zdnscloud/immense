## 存储类型

pkg/eventhandler/eventhandler.go

```
func New(cli client.Client) *HandlerManager {
        return &HandlerManager{
                handlers: []Handler{
                        lvm.New(cli),
                        ceph.New(cli),
                },
        }
}
```
## 逻辑
### lvm说明
目录结构
```
└── lvm
    ├── check.go      #不再使用
    ├── create.go     #创建存储集群
    ├── delete.go     #删除存储集群
    ├── lvm.go        #主文件
    ├── status.go     #更新存储集群状态
    ├── template.go   #模板文件
    ├── update.go     #更新存储集群配置
    └── yaml.go       #配置文件
```

- 创建  
  1. 给节点增加labels和annotations
  2. 部署Lvmd
  3. 初始化磁盘
     - 检查Volume Group是否已经存在
     - 检查磁盘是否有分区和文件系统，如果有则强制格式化磁盘
     - 创建Physical volume
     - 创建Volume Group
    >
    >  注：如果创建Volume Group之前不存在，则直接vgcreate。如果已经存在，则进行vgextend
  4. 部署CSI
  5. gorouting循环检查lvm的运行及磁盘空间并更新cluster状态（频率60秒）   
  
- 更新  
  1. 对比更新前后的配置，确定删除的主机、增加的主机、删除的磁盘、增加的磁盘
  2. 对上面4种配置进行分别处理
  > 
  > 如果删除前只有一块磁盘组成Volume Group，则直接vgremove。如果是有多块磁盘组成Volume Group，则进行vgreduce
  
  > 如果有Pod正在使用这个Volume Group，则Volume Group的操作将会失败
     
- 删除  
  1. 删除CSI
  2. 格式化磁盘
  3. 删除Lvmd
  4. 删除节点的labels和annotations
  
  
  ### ceph说明
目录结构
```
├── ceph
│   ├── ceph.go     #主文件
│   ├── client      #包装ceph命令
│   ├── config      #为ceph集群创建configmap,secret,headless-service
│   ├── create.go   #创建存储集群
│   ├── delete.go   #删除存储集群
│   ├── fscsi       #CSI相关
│   ├── global      #全局变量配置
│   ├── mds         #ceph组件mds
│   ├── mgr         #ceph组件mgr         
│   ├── mon         #ceph组件mon
│   ├── osd         #ceph组件osd
│   ├── status      #更新存储集群状态
│   ├── update.go   #更惨存储集群配置
│   ├── util        #常用工具函数
│   └── zap         #初始化磁盘
```
- 创建  
  1. 给节点增加labels和annotations
  2. 创建ceph集群
      1. 获取k8s集群Pod地址段（当前固定为10.42.0.0/16）
      2. 随机生成uuid, adminkey, monkey
      3. 根据磁盘个数设置副本数（默认为2）
      4. 根据前面4步的配置
          - 创建configmap保存ceph集群配置文件，用于后面启动ceph组件挂载使用
          - 创建无头服务，用于后面ceph组件连接mon
          - 创建secret，保存账户和密钥，用于后面storageclass使用
      5. 保存ceph集群配置到本地
      6. 启动mon并等待其全部运行
      7. 启动mgr
      8. 启动osd（先调用zap对磁盘进行清理）
      9. 启动mds
  3. 部署CSI
  4. 启动3个gorouting
     - 循环检查ceph集群中是否有异常的osd，如果有就remove，等待集群数据恢复
     - 循环检查ceph集群中是否有异常的mon，如果有就remove
     - 循环检查ceph的运行及磁盘空间并更新cluster状态（频率60秒）    
- 更新  
  1. 对比更新前后的配置，确定删除的主机、增加的主机、删除的磁盘、增加的磁盘
  2. 实际上就是增加/删除osd组件Pod
  3. 删除后增加labels和annotations
- 删除  
  1. 删除CSI
  2. 删除Ceph集群
     1. 删除mds
     2. 删除osd（后调用zap对磁盘进行清理）
     3. 删除mgr
     4. 删除mon
     5. 删除本地ceph配置文件
  3. 删除configmap,secret,service
  3. 删除节点的labels和annotations


