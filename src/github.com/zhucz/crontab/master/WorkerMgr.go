package master

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"task_scheduler/src/github.com/zhucz/crontab/common"
	"time"
)

// WorkerMgr 复制获取/cron/worker/下的kv
type WorkerMgr struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

var G_workerMgr *WorkerMgr

// ListWorkers 获取在线worker列表
func (workerMgr *WorkerMgr) ListWorkers() (workerArr []string, err error) {
	var (
		getResp *clientv3.GetResponse
	)
	workerArr = make([]string, 0)
	if getResp, err = workerMgr.kv.Get(context.TODO(), common.JobWorkerDir, clientv3.WithPrefix()); err != nil {
		return
	}

	// 解析每个节点的IP
	for _, kv := range getResp.Kvs {
		//key: /cron/workers/169.254.218.170
		key := string(kv.Key)

		workerArr = append(workerArr, common.ExtractWorkerIP(key))
	}
	return
}

func InitWorkerMgr() (err error) {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
	)

	config = clientv3.Config{
		Endpoints:   G_config.EtcdHosts,
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond,
	}

	if client, err = clientv3.New(config); err != nil {
		return
	}

	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	G_workerMgr = &WorkerMgr{
		client: client,
		kv:     kv,
		lease:  lease,
	}
	return
}
