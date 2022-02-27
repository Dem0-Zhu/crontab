package master

import (
	"context"
	"encoding/json"
	"fmt"
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
}

func InitJobMgr() (err error) {
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

	G_jobMgr = &JobMgr{
		client: client,
		kv:     kv,
		lease:  lease,
	}
	return
}

func (jobMgr *JobMgr) PutJob(job *common.Job) (oldJob *common.Job, err error) {
	// 把任务保存到/cron/jobs/任务名 ---> json
	var (
		jobKey    string
		jobValue  []byte
		putResp   *clientv3.PutResponse
		oldJobObj common.Job
	)
	jobKey = common.JobSaveDir + job.Name
	if jobValue, err = json.Marshal(job); err != nil {
		return
	}
	// 保存到etcd
	if putResp, err = jobMgr.kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV()); err != nil {
		return
	}
	// 如果更新，返回旧值
	if putResp.PrevKv != nil {
		// 对旧值做一个反序列化
		if err = json.Unmarshal(putResp.PrevKv.Value, &oldJobObj); err != nil {
			err = nil
			return
		}
		oldJob = &oldJobObj
	}
	return
}

func (jobMgr *JobMgr) DeleteJob(name string) (oldJob *common.Job, err error) {
	var (
		jobKey    string
		delResp   *clientv3.DeleteResponse
		oldJobObj common.Job
	)
	jobKey = common.JobSaveDir + name
	if delResp, err = jobMgr.kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV()); err != nil {
		return
	}
	if len(delResp.PrevKvs) != 0 {
		if err = json.Unmarshal(delResp.PrevKvs[0].Value, &oldJobObj); err != nil {
			// 允许旧值解析失败
			err = nil
			return
		}
		oldJob = &oldJobObj
	}
	return
}

func (jobMgr *JobMgr) GetJob(name string) (job *common.Job, err error) {
	var (
		jobKey  string
		getResp *clientv3.GetResponse
		jobObj  common.Job
	)
	jobKey = common.JobSaveDir + name
	fmt.Println(jobKey)
	if getResp, err = jobMgr.kv.Get(context.TODO(), jobKey); err != nil {
		return
	}
	fmt.Println(getResp)
	if len(getResp.Kvs) != 0 {
		if err = json.Unmarshal(getResp.Kvs[0].Value, &jobObj); err != nil {
			// 允许旧值解析失败
			err = nil
			return
		}
		job = &jobObj
	}
	return
}

func (jobMgr *JobMgr) ListJobs() (jobList map[string]common.Job, err error) {
	var (
		jobKey  string
		getResp *clientv3.GetResponse
		job     common.Job
	)
	jobKey = common.JobSaveDir

	if getResp, err = jobMgr.kv.Get(context.TODO(), jobKey, clientv3.WithPrefix()); err != nil {
		return
	}

	jobList = make(map[string]common.Job, 0)
	if len(getResp.Kvs) != 0 {
		for _, kvPair := range getResp.Kvs {
			if err = json.Unmarshal(kvPair.Value, &job); err != nil {
				err = nil
				// 容忍存在反序列化失败
				continue
			}
			jobList[string(kvPair.Key)] = job
		}
	}
	return
}

func (jobMgr *JobMgr) KillJob(name string) (err error) {
	var (
		killerJob      string
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseId        clientv3.LeaseID
	)
	killerJob = common.JobKillerDir + name
	if leaseGrantResp, err = jobMgr.lease.Grant(context.TODO(), 1); err != nil {
		return
	}

	leaseId = leaseGrantResp.ID

	if _, err = jobMgr.kv.Put(context.TODO(), killerJob, "", clientv3.WithLease(leaseId)); err != nil {
		return
	}
	return
}
