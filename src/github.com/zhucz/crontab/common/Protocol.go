package common

import (
	"context"
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
	Job *Job // 调度的任务信息
	Expr *cronexpr.Expression // 解析好的任务表达式
	NextTime time.Time // 下次调度时间

}

// JobExecuteInfo 任务执行状态
type JobExecuteInfo struct {
	Job *Job
	PlanTime time.Time // 理论上的调度时间
	RealTime time.Time // 实际上的调度时间
	CancelCtx context.Context // 任务command的context
	CancelFunc context.CancelFunc // 用于取消command执行的cancel函数
}

// JobExecuteResult 任务执行结果
type JobExecuteResult struct {
	ExecuteInfo *JobExecuteInfo
	Output []byte
	Err error
	StartTime time.Time
	EndTime time.Time
}

// JobLog 任务执行日志
type JobLog struct {
	JobName string `json:"job_name" bson:"job_name"`
	Command string `json:"command" bson:"command"`
	Err string `json:"err" bson:"err"`
	Output string `json:"output" bson:"output"`
	PlanTime int64 `json:"plan_time" bson:"plan_time"` // 计划开始时间
	ScheduleTime int64 `json:"schedule_time" bson:"schedule_time"` // 实际调度时间
	StartTime int64 `json:"start_time" bson:"start_time"` // 任务执行开始时间
	EndTime int64 `json:"end_time" bson:"end_time"` // 任务执行结束时间
}

// LogBatch 日志批次
type LogBatch struct {
	Logs []interface{}
}

type JobLogFilter struct {
	JobName string `bson:"job_name"`
}

type SortLogByStartTime struct {
	StartTime int `bson:"start_time"`
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

// ExtractKillerName 从/cron/killer/job中提取任务名
func ExtractKillerName(killerKey string) string {
	return strings.TrimPrefix(killerKey, JobKillerDir)
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

// BuildJobExecuteInfo 构造执行状态信息
func BuildJobExecuteInfo(jobSchedulePlan *JobSchedulePlan) (jobExecuteInfo *JobExecuteInfo) {
	jobExecuteInfo = &JobExecuteInfo{
		Job: jobSchedulePlan.Job,
		PlanTime: jobSchedulePlan.NextTime,
		RealTime: time.Now(),
	}
	jobExecuteInfo.CancelCtx, jobExecuteInfo.CancelFunc = context.WithCancel(context.TODO())
	return
}