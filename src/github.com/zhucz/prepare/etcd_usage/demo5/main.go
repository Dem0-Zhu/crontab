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
		delResp *clientv3.DeleteResponse
	)
	config = clientv3.Config{
		Endpoints:   []string{"192.168.5.16:2379"},
		DialTimeout: 5 * time.Second,
	}

	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}

	kv = clientv3.NewKV(client)

	if delResp, err = kv.Delete(context.TODO(), "cron/job/job3", clientv3.WithPrevKV()); err != nil {
		fmt.Println(err)
	} else {
		if len(delResp.PrevKvs) != 0 {
			for _, kvPair := range delResp.PrevKvs {
				fmt.Println(kvPair)
			}
		}
	}

}
