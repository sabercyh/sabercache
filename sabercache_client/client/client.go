package client

import (
	"context"
	"fmt"
	"log"
	"sabercache_client/consistenthash"
	pb "sabercache_client/sabercachepb"
	"sabercache_client/util"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Client struct {
	consistenthash *consistenthash.Consistency
	peers          []string
}

func NewClient() *Client {
	peers := DiscoverPeers()
	log.Println(peers)
	c := &Client{peers: peers}
	c.consistenthash = consistenthash.New(util.Replicas, nil)
	c.consistenthash.Register(peers)
	return c
}
func (c *Client) Get(key string) ([]byte, error) {
	cli, err := clientv3.New(defaultEtcdConfig)
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	peer := c.consistenthash.GetPeer(key)
	conn, err := EtcdDial(cli, peer)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	grpcClient := pb.NewSaberCacheClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := grpcClient.Get(ctx, &pb.GetRequest{
		Key: key,
	})
	if err != nil {
		return nil, fmt.Errorf("could not get %s from peer %s", key, peer)
	}
	log.Printf("get %s from %s\n", key, peer)
	return resp.GetValue(), nil
}

func (c *Client) GetAll() (respKV []*pb.KeyValue, err error) {
	cli, err := clientv3.New(defaultEtcdConfig)
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	for _, peer := range c.peers {
		conn, err := EtcdDial(cli, peer)
		if err != nil {
			return nil, err
		}
		defer conn.Close()
		grpcClient := pb.NewSaberCacheClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		resp, err := grpcClient.GetAll(ctx, &pb.GetAllRequest{})
		if err != nil {
			return nil, fmt.Errorf("could not getall from peer %s", peer)
		}
		log.Printf("getall from %s\n", peer)
		respKV = append(respKV, resp.Kv...)
	}

	return
}

func (c *Client) Set(key string, value []byte, ttl int64) (bool, error) {
	cli, err := clientv3.New(defaultEtcdConfig)
	if err != nil {
		return false, err
	}
	defer cli.Close()
	peer := c.consistenthash.GetPeer(key)
	conn, err := EtcdDial(cli, peer)
	if err != nil {
		return false, err
	}
	defer conn.Close()
	grpcClient := pb.NewSaberCacheClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := grpcClient.Set(ctx, &pb.SetRequest{
		Key:   key,
		Value: value,
		Ttl:   ttl,
	})
	if err != nil {
		return false, fmt.Errorf("could not set %s to peer %s", key, peer)
	}
	log.Printf("set %s to %s\n", key, peer)
	return resp.Ok, nil
}

func (c *Client) TTL(key string) (int64, error) {
	cli, err := clientv3.New(defaultEtcdConfig)
	if err != nil {
		return -2, err
	}
	defer cli.Close()
	peer := c.consistenthash.GetPeer(key)
	conn, err := EtcdDial(cli, peer)
	if err != nil {
		return -2, err
	}
	defer conn.Close()
	grpcClient := pb.NewSaberCacheClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := grpcClient.TTL(ctx, &pb.TTLRequest{
		Key: key,
	})
	if err != nil {
		return -2, fmt.Errorf("could not set %s to peer %s", key, peer)
	}
	log.Printf("get ttl %s from %s\n", key, peer)
	return resp.Ttl, nil
}
