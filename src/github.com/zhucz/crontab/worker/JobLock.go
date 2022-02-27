package worker

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"task_scheduler/src/github.com/zhucz/crontab/common"
)

// JobLock 分布式锁
type JobLock struct {
	kv    clientv3.KV
	lease clientv3.Lease

	jobName    string // 任务名
	leaseID    clientv3.LeaseID
	cancelFunc context.CancelFunc // 用于终止续租
	isLocked   bool
}

// InitJobLock 初始化一把锁
func InitJobLock(jobName string, kv clientv3.KV, lease clientv3.Lease) (jobLock *JobLock) {
	jobLock = &JobLock{
		kv:      kv,
		lease:   lease,
		jobName: jobName,
	}
	return
}

// TryLock 尝试上锁
func (jobLock *JobLock) TryLock() (err error) {

	var (
		leaseGrantResp    *clientv3.LeaseGrantResponse
		cancelCtx         context.Context
		cancelFunc        context.CancelFunc
		leaseID           clientv3.LeaseID
		keepAliveRespChan <-chan *clientv3.LeaseKeepAliveResponse
		keepAliveResp     *clientv3.LeaseKeepAliveResponse
		txnResp           *clientv3.TxnResponse
		txn               clientv3.Txn
		lockKey           string
	)

	// 1. 创建租约，节点宕机，租约到期自动释放
	if leaseGrantResp, err = jobLock.lease.Grant(context.TODO(), 5); err != nil {
		return err
	}
	// 用于取消自动续租
	cancelCtx, cancelFunc = context.WithCancel(context.TODO())

	// 获取租约id
	leaseID = leaseGrantResp.ID

	// 2. 自动续租
	if keepAliveRespChan, err = jobLock.lease.KeepAlive(cancelCtx, leaseID); err != nil {
		goto FAIL
	}
	// 3. 处理续租应答的协程
	go func() {
		for {
			select {
			case keepAliveResp = <-keepAliveRespChan: // 自动续租应答
				if keepAliveResp == nil { // 说明自动续租被取消掉
					goto END
				}
			}
		}
	END:
	}()

	// 4. 创建事务txn
	txn = jobLock.kv.Txn(context.TODO())
	// 锁路径
	lockKey = common.JobLockDir + jobLock.jobName
	// 5.事务抢锁    lockKey的创建revision等于0（未创建）,则抢锁成功
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(lockKey, "", clientv3.WithLease(leaseID))).
		Else(clientv3.OpGet(lockKey))

	// 提交事务
	if txnResp, err = txn.Commit(); err != nil {
		goto FAIL
	}
	// 6. 成功返回；失败释放租约
	if !txnResp.Succeeded {
		// 锁被占用
		err = common.ErrLockAlreadyRequired
		goto FAIL
	}

	// 抢锁成功
	jobLock.leaseID = leaseID
	jobLock.cancelFunc = cancelFunc
	jobLock.isLocked = true
	return
FAIL:
	// 取消自动续租
	cancelFunc()
	// 释放租约
	jobLock.lease.Revoke(context.TODO(), leaseID)
	return
}

// UnLock 释放锁
func (jobLock *JobLock) UnLock() {
	if jobLock.isLocked {
		jobLock.cancelFunc()
		jobLock.lease.Revoke(context.TODO(), jobLock.leaseID)
	}

}
