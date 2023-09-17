package registry

import (
	"context"
	"strings"

	clientv3 "go.etcd.io/etcd/client/v3"
)

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
