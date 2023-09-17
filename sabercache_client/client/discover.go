package client

import (
	"context"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
)

var (
	defaultEtcdConfig = clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	}
)

func EtcdDial(c *clientv3.Client, service string) (*grpc.ClientConn, error) {
	etcdResolver, err := resolver.NewBuilder(c)
	if err != nil {
		return nil, err
	}
	return grpc.Dial(
		"etcd:///"+"sabercache/"+service,
		grpc.WithResolvers(etcdResolver),
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
}
func DiscoverPeers() (peers []string) {
	cli, err := clientv3.New(defaultEtcdConfig)
	rangeResp, err := cli.Get(context.TODO(), "sabercache/", clientv3.WithPrefix())
	if err != nil {
		panic(err)
	}
	for _, v := range rangeResp.Kvs {
		service := string(v.Key)
		peer := strings.Split(service, "/")[1]
		peers = append(peers, peer)
	}
	return
}
