package main

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

func main() {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		err     error
		kv      clientv3.KV
		putResp *clientv3.PutResponse
	)
	config = clientv3.Config{
		Endpoints:   []string{"192.168.5.16:2379"},
		DialTimeout: 5 * time.Second,
	}

	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}

	// kv用于读写etcd的键值对，kv内部支持重试
	kv = clientv3.NewKV(client)
	if putResp, err = kv.Put(context.TODO(), "cron/job/job1", "hello"); err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("putResp.Header.Revision = %v", putResp.Header.Revision)
	}

	// 查看key对应的上一个value
	if putResp, err = kv.Put(context.TODO(), "cron/job/job1", "bye", clientv3.WithPrevKV()); err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("putResp.Header.Revision = %v", putResp.Header.Revision)
		if putResp.PrevKv != nil {
			fmt.Printf("putResp.PrevKv.Value = %v", string(putResp.PrevKv.Value))
		}
	}

}
