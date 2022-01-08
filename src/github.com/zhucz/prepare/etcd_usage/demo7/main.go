package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

func main() {
	var (
		config         clientv3.Config
		client         *clientv3.Client
		err            error
		kv             clientv3.KV
		getResp        *clientv3.GetResponse
		startReversion int64
		watchChan      <-chan clientv3.WatchResponse
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

	go func() {
		for {
			kv.Put(context.TODO(), "cron/jobs/job7", "job7...")

			kv.Delete(context.TODO(), "cron/jobs/job7")

			time.Sleep(1 * time.Second)
		}
	}()

	// 先get到当前值，并监听后续变化
	if getResp, err = kv.Get(context.TODO(), "cron/jobs/job7"); err != nil {
		fmt.Println(err)
		return
	}
	if len(getResp.Kvs) == 0 {
		fmt.Println("当前值为：", string(getResp.Kvs[0].Value))
	}

	// Revision为当前etcd集群事务id，单调递增，从Revision + 1开始监控；
	startReversion = getResp.Header.Revision + 1

	watch := clientv3.NewWatcher(client)

	ctx, cancelFunc := context.WithCancel(context.TODO())
	time.AfterFunc(5*time.Second, func() {
		cancelFunc()
	})
	watchChan = watch.Watch(ctx, "cron/jobs/job7", clientv3.WithRev(startReversion))

	for watch := range watchChan {
		for _, event := range watch.Events {
			switch event.Type {
			case mvccpb.PUT:
				fmt.Println("修改：", event.Kv.Value, "Reversion: ", event.Kv.CreateRevision, event.Kv.ModRevision)
			case mvccpb.DELETE:
				fmt.Println("删除了：", event.Kv.Value, "Revension: ", event.Kv.CreateRevision, event.Kv.ModRevision)
			}
		}
	}
}
