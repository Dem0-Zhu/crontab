package main

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

// 乐观锁
/*
实现：lease实现锁自动过期；op操作；txn事务
上锁（创建租约、自动续租、拿着租约去抢占一个key）；
处理事务；
释放锁（取消自动续租、释放租约）；
*/
func main() {
	var (
		err                error
		client             *clientv3.Client
		lgResp             *clientv3.LeaseGrantResponse
		leaseKeepAliveResp <-chan *clientv3.LeaseKeepAliveResponse
		txnResp            *clientv3.TxnResponse
	)
	config := clientv3.Config{
		Endpoints:   []string{"192.168.5.16:2379"},
		DialTimeout: 5 * time.Second,
	}
	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}

	//创建租约
	lease := clientv3.NewLease(client)
	// 创建一个五秒的租约
	if lgResp, err = lease.Grant(context.TODO(), 5); err != nil {
		fmt.Println(err)
		return
	}

	leaseID := lgResp.ID

	// 自动续租
	ctx, cancelFunc := context.WithCancel(context.TODO())

	// 确保函数结束时，停止续租
	defer func() {
		cancelFunc()
		lease.Revoke(context.TODO(), leaseID)
	}()

	if leaseKeepAliveResp, err = lease.KeepAlive(ctx, leaseID); err != nil {
		fmt.Println(err)
		return
	}

	// 处理续租应答
	go func() {
		for {
			select {
			case resp := <-leaseKeepAliveResp:
				if resp == nil {
					fmt.Println("租约失效...")
					goto END
				} else {
					fmt.Println("收到租约应答...", resp.ID)
				}
			}
		}
	END:
	}()

	// 抢key
	kv := clientv3.NewKV(client)
	// 创建一个事务
	txn := kv.Txn(context.TODO())
	txn.If(clientv3.Compare(clientv3.CreateRevision("cron/jobs/job9"), "=", 0)).
		Then(clientv3.OpPut("cron/jobs/job9", "xxx", clientv3.WithLease(leaseID))).
		Else(clientv3.OpGet("cron/jobs/job9"))

	// 提交事务
	if txnResp, err = txn.Commit(); err != nil {
		fmt.Println(err)
		return
	}
	//判断是否抢到锁
	if !txnResp.Succeeded {
		fmt.Printf("锁被%v占用", txnResp.Responses[0].GetResponseRange().Kvs[0].Value)
		return
	}

	// 处理业务
	fmt.Println("业务处理...")
	time.Sleep(5 * time.Second)

	// 释放锁

}
