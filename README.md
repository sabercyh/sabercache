# 仿Memcached分布式缓存系统
## 系统结构
![系统结构图](/assets/System_structure.png)
## 系统介绍
* 支持多种缓存淘汰策略，如FIFO，LRU，LFU
* 支持设置缓存过期时间，通过惰性删除和定期删除组合的方式删除过期数据
* 缓存未命中时采用singleflight实现数据加载，防缓存穿透
* 系统在客户端通过一致性哈希实现负载均衡
* 使用etcd作为服务注册中心，客户端和服务端节点间通过gRPC实现服务调用
## 系统使用
```
cd sabercache_server/server && go run main.go
cd sabercache_client && go run main.go
go run main.go --tcpAddr 指定地址
```
## 系统命令
```
set k1 v1
get k1

set k2 100 v2 
get v2

ttl v2

getall

save

exit
```
## TODO
* 客户端定时监听服务端节点变化，并调整服务端节点缓存
* 服务端正常下线