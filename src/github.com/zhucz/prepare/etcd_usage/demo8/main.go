package main

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

func main() {
	var (
		client *clientv3.Client
		opResp clientv3.OpResponse
		err    error
	)
	config := clientv3.Config{
		Endpoints:   []string{"192.168.5.16:2379"},
		DialTimeout: 5 * time.Second,
	}

	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}

	kv := clientv3.NewKV(client)
	putOp := clientv3.OpPut("cron/jobs/job8", "123")
	if opResp, err = kv.Do(context.TODO(), putOp); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(opResp.Put().Header.Revision)

	getOp := clientv3.OpGet("cron/jobs/job8")
	if opResp, err = kv.Do(context.TODO(), getOp); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(opResp.Get().Kvs[0].ModRevision)
	fmt.Println(opResp.Get().Kvs[0].Value)

}
