package worker

import (
	"context"
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
		startTime := time.Now()
		// 执行shell命令
		cmd := exec.CommandContext(context.TODO(), "C:\\Windows\\System32\\bash.exe", "-c", info.Job.Command)
		// 执行并捕获输出
		output, err := cmd.CombinedOutput()
		eEndTime := time.Now()

		// 将执行结果返回给Scheduler
		result := &common.JobExecuteResult{
			ExecuteInfo: info,
			Output: output,
			Err: err,
			StartTime: startTime,
			EndTime: eEndTime,
		}
		GScheduler.PushJobResult(result)
	}()
}

func InitExecutor() (err error) {
	G_Executor = &Executor{}
	return
}
