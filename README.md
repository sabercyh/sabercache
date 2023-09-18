# 分布式缓存系统
## 系统结构
![系统结构图](/assets/System_structure.png)
## Start
```
go run main.go --tcpAddr 指定地址
```
## Command
```
set k1 v1
get k1

set k2 100 v2 
get v2
ttl v2

getall
```