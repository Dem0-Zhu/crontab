package worker

import (
	"math/rand"
	"os/exec"
	"task_scheduler/src/github.com/zhucz/crontab/common"
	"time"
)

type Executor struct {
}

var (
	G_Executor *Executor
)

func (executor *Executor) ExecutorJob(info *common.JobExecuteInfo) {
	go func() {
		var (
			err    error
			output []byte
		)
		// 初始化锁
		jobLock := G_jobMgr.CreatJobLock(info.Job.Name)

		startTime := time.Now()

		// 随机睡眠0~1s, 尽量让所有的worker都有机会抢到锁
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		err = jobLock.TryLock()
		defer jobLock.UnLock()

		result := &common.JobExecuteResult{
			ExecuteInfo: info,
			StartTime:   startTime,
		}
		if err != nil {
			result.Err = err
			result.EndTime = time.Now()
		} else {
			startTime = time.Now()
			// 执行shell命令
			cmd := exec.CommandContext(info.CancelCtx, "C:\\Windows\\System32\\bash.exe", "-c", info.Job.Command)
			// 执行并捕获输出
			output, err = cmd.CombinedOutput()
			endTime := time.Now()
			result.Err = err
			result.StartTime = startTime
			result.EndTime = endTime
			result.Output = output
		}
		// 将执行结果返回给Scheduler
		GScheduler.PushJobResult(result)
	}()
}

func InitExecutor() (err error) {
	G_Executor = &Executor{}
	return
}
