package main

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

// 租约
func main() {
	var (
		config      clientv3.Config
		client      *clientv3.Client
		err         error
		lgResp      *clientv3.LeaseGrantResponse
		putResp     *clientv3.PutResponse
		getResp     *clientv3.GetResponse
		lkaRespChan <-chan *clientv3.LeaseKeepAliveResponse
	)
	config = clientv3.Config{
		Endpoints:   []string{"192.168.5.16:2379"},
		DialTimeout: 5 * time.Second,
	}

	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}

	lease := clientv3.NewLease(client)
	// 申请一个十秒的租约
	if lgResp, err = lease.Grant(context.TODO(), 10); err != nil {
		fmt.Println(err)
		return
	}

	// 续租 每秒进行一次续租,第五秒取消续租
	ctx, _ := context.WithTimeout(context.TODO(), 5*time.Second)
	if lkaRespChan, err = lease.KeepAlive(ctx, lgResp.ID); err != nil {
		fmt.Println(err)
		return
	}

	// 处理
	go func() {
		for {
			select {
			case resp := <-lkaRespChan:
				if resp == nil {
					fmt.Println("租约失效...")
					goto END
				} else {
					fmt.Println("收到自动续约应答: ", resp.ID)
				}
			}
		}
	END:
	}()

	kv := clientv3.NewKV(client)
	if putResp, err = kv.Put(context.TODO(), "cron/lock/job1", "...", clientv3.WithLease(lgResp.ID)); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("写入成功", putResp.Header.Revision)

	for {
		if getResp, err = kv.Get(context.TODO(), "cron/lock/job1"); err != nil {
			fmt.Println(err)
			return
		}
		if getResp.Count == 0 {
			fmt.Println("key 过期了")
			break
		}
		fmt.Println("key没过期：", getResp.Kvs)
		time.Sleep(2 * time.Second)
	}

}
