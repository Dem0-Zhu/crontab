package worker

import (
	"context"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"task_scheduler/src/github.com/zhucz/crontab/common"
	"time"
)

var (
	G_jobMgr *JobMgr
)

// JobMgr 任务管理器
type JobMgr struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
	watcher clientv3.Watcher
}

func InitJobMgr() (err error) {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
		watcher clientv3.Watcher
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
	watcher = clientv3.NewWatcher(client)

	G_jobMgr = &JobMgr{
		client: client,
		kv:     kv,
		lease:  lease,
		watcher: watcher,
	}

	G_jobMgr.watchJobs()

	return
}

func (jobMgr *JobMgr) watchJobs() (err error) {
	var (
		getResp *clientv3.GetResponse
		kvPair *mvccpb.KeyValue
		job *common.Job
		watchStartRevision int64
		watchChan clientv3.WatchChan
		watchResp clientv3.WatchResponse
		watchEvent *clientv3.Event
		jobEvent *common.JobEvent
	)
	// 1. get到/cron/jobs/目录下的所有任务，并且获知当前键值对的revision
	if getResp, err = jobMgr.kv.Get(context.TODO(), common.JobSaveDir, clientv3.WithPrefix()); err != nil {
		return
	}

	// 遍历当前任务
	for _, kvPair = range getResp.Kvs {
		// 反序列化json得到job
		if job, err = common.UnpackJob(kvPair.Value); err == nil {
			jobEvent = common.BuildJobEvent(common.JobEventPut, job)
			// 同步给调度协程（scheduler）
			GScheduler.PushJobEvent(jobEvent)
		}
	}

	// 2. 从该revision向后监听变化事件
	go func() { // 监听协程
		// 从get时刻的后续版本开始监听
		watchStartRevision = getResp.Header.Revision + 1
		// 启动监听/cron/jobs/目录的后续变化
		watchChan = jobMgr.watcher.Watch(context.TODO(),common.JobSaveDir,clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())
		for watchResp = range watchChan {
			// 每个watchResp有可能包含多个Event
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT:
					// 更新
					if job, err = common.UnpackJob(watchEvent.Kv.Value); err != nil {
						continue
					}
					//构造一个Event事件
					jobEvent = common.BuildJobEvent(common.JobEventPut, job)
				case mvccpb.DELETE:
					// delete /cron/jobs/job10, 需要得到job10
					jobName := common.ExtractJobName(string(watchEvent.Kv.Key))
					//构造一个Event事件
					jobEvent = common.BuildJobEvent(common.JobEventDelete, &common.Job{Name:jobName})
				}
				// 把变化推送给scheduler
				GScheduler.PushJobEvent(jobEvent)

			}
		}
	}()

	return nil
}
