package common

import (
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"strings"
	"time"
)

type Job struct {
	Name     string `json:"name"`      // 任务名
	Command  string `json:"command"`   // shell命令
	CronExpr string `json:"cron_expr"` // cron表达式
}

// Response HTTP接口应答
type Response struct {
	Errno int         `json:"errno"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}

// JobEvent 变化事件
type JobEvent struct {
	EventType int
	Job *Job
}

// JobSchedulePlan 调度计划
type JobSchedulePlan struct {
	Job *Job //调度的任务信息
	Expr *cronexpr.Expression // 解析好的任务表达式
	NextTime time.Time // 下次调度时间

}

// JobExecuteInfo 任务执行状态
type JobExecuteInfo struct {
	Job *Job
	PlanTime time.Time // 理论上的调度时间
	RealTime time.Time // 实际上的调度时间
}

// JobExecuteResult 任务执行结果
type JobExecuteResult struct {
	ExecuteInfo *JobExecuteInfo
	Output []byte
	Err error
	StartTime time.Time
	EndTime time.Time
}

// BuildResponse 应答方法
func BuildResponse(errno int, msg string, data interface{}) (resp []byte, err error) {
	response := Response{
		Errno: errno,
		Msg:   msg,
		Data:  data,
	}
	resp, err = json.Marshal(response)
	return
}

// UnpackJob 反序列化Job
func UnpackJob(value []byte) (ret *Job, err error) {
	var (
		job *Job
	)
	job = &Job{}
	if err = json.Unmarshal(value, job); err != nil {
		return nil, err
	}
	ret = job
	return ret, nil
}

// ExtractJobName 从etcd的key中提取任务名
func ExtractJobName(jobKey string) string {
	return strings.TrimPrefix(jobKey, JobSaveDir)
}

func BuildJobEvent(eventType int, job *Job) *JobEvent {
	return &JobEvent{
		EventType: eventType,
		Job: job,
	}
}

// BuildJobSchedulePlan 构造任务执行计划
func BuildJobSchedulePlan(job *Job) (jobSchedulePlan *JobSchedulePlan, err error) {
	var (
		expr *cronexpr.Expression
	)
	// 解析任务cron表达式
	if expr, err = cronexpr.Parse(job.CronExpr); err != nil {
		return nil, err
	}
	// 生成调度计划
	jobSchedulePlan = &JobSchedulePlan{
		Job: job,
		Expr: expr,
		NextTime: expr.Next(time.Now()),
	}
	return
}

//
func BuildJobExecuteInfo(jobSchedulePlan *JobSchedulePlan) (jobExecuteInfo *JobExecuteInfo) {
	jobExecuteInfo = &JobExecuteInfo{
		Job: jobSchedulePlan.Job,
		PlanTime: jobSchedulePlan.NextTime,
		RealTime: time.Now(),
	}
	return
}