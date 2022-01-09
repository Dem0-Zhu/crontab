package worker

import (
	"fmt"
	"task_scheduler/src/github.com/zhucz/crontab/common"
	"time"
)

// Scheduler 任务调度, 不停地循环检查任务是否到期
type Scheduler struct {
	jobEventChan chan *common.JobEvent              // etcd任务事件队列
	jobPlanTable map[string]*common.JobSchedulePlan // 任务调度计划表：任务名称 --》 调度计划
	jobExecutingTable map[string]*common.JobExecuteInfo // 正在执行的任务
	jobResultChan chan *common.JobExecuteResult // 保存任务执行完成的结果
}

var (
	GScheduler *Scheduler
)

// HandleJobEvent 处理监听到的任务事件， 更新内存中的任务表，使得与etcd中任务信息保持一致
func (scheduler *Scheduler) HandleJobEvent(jobEvent *common.JobEvent) {
	var (
		jobSchedulePlan *common.JobSchedulePlan
		err             error
	)
	switch jobEvent.EventType {
	case common.JobEventPut:
		// 保存/更新任务事件
		if jobSchedulePlan, err = common.BuildJobSchedulePlan(jobEvent.Job); err != nil {
			return
		}
		GScheduler.jobPlanTable[jobEvent.Job.Name] = jobSchedulePlan
	case common.JobEventDelete:
		// 删除任务事件
		if _, ok := GScheduler.jobPlanTable[jobEvent.Job.Name]; ok {
			delete(GScheduler.jobPlanTable, jobEvent.Job.Name)
		}
	}
}

// HandleJobResult 处理任务结果
func (scheduler *Scheduler) HandleJobResult(jobResult *common.JobExecuteResult) {
	if _, ok := scheduler.jobExecutingTable[jobResult.ExecuteInfo.Job.Name]; ok {
		delete(scheduler.jobExecutingTable, jobResult.ExecuteInfo.Job.Name)
	}

	fmt.Println("任务执行完成：", jobResult.ExecuteInfo.Job.Name, string(jobResult.Output), jobResult.StartTime, jobResult.EndTime, jobResult.Err)
}

func (scheduler *Scheduler)TryStartJob(jobSchedulePlan *common.JobSchedulePlan)  {
	// 假如任务每一秒钟被调度一次，但会执行1分钟，所以一分钟会被调度60次，这里保证正在运行的任务只会被调度一次
	if _, ok := scheduler.jobExecutingTable[jobSchedulePlan.Job.Name]; ok {
		// 任务已经被调度
		return
	}

	// 构建执行状态信息, 保存执行状态
	scheduler.jobExecutingTable[jobSchedulePlan.Job.Name] = common.BuildJobExecuteInfo(jobSchedulePlan)

	// 执行任务
	G_Executor.ExecutorJob(scheduler.jobExecutingTable[jobSchedulePlan.Job.Name])
	fmt.Println("执行任务：", jobSchedulePlan.Job.Name, scheduler.jobExecutingTable[jobSchedulePlan.Job.Name].PlanTime, scheduler.jobExecutingTable[jobSchedulePlan.Job.Name].RealTime)
}

// TrySchedule 遍历任务， 执行过期任务，返回下次执行的最小时间间隔
func (scheduler *Scheduler) TrySchedule() (scheduleAfter time.Duration) {
	var (
		nearTime *time.Time
	)
	if len(scheduler.jobPlanTable) == 0 {
		time.Sleep(1 * time.Second)
		return
	}

	now := time.Now()
	for _, jobPlan := range scheduler.jobPlanTable {
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {
			// todo: 尝试执行任务
			scheduler.TryStartJob(jobPlan)
			jobPlan.NextTime = jobPlan.Expr.Next(now)
		}

		// 正常情况需要轮训遍历planTable， 这边加了一个小优化
		// 统计最近一个要过期的任务
		if nearTime == nil || jobPlan.NextTime.Before(*nearTime) {
			nearTime = &jobPlan.NextTime
		}
	}

	// 下次调度时间间隔: 距离最近的任务 - 当前时间
	scheduleAfter = (*nearTime).Sub(now)
	return
}

// 调度协程
func (scheduler *Scheduler) schedulerLoop() {
	var (
		scheduleAfter time.Duration
		scheduleTimer *time.Timer
	)

	// 初始化一次
	scheduleAfter = scheduler.TrySchedule()
	// 定义一个调度的定时器
	scheduleTimer = time.NewTimer(scheduleAfter)

	for {
		select {
		case jobEvent := <-scheduler.jobEventChan: // 监听任务变化事件
			// 对内存中维护的列表做增删改查
			GScheduler.HandleJobEvent(jobEvent)
		case <-scheduleTimer.C: // 最近的任务到期了
		case jobResult := <-scheduler.jobResultChan:
			scheduler.HandleJobResult(jobResult)
		}
		// 调度一次任务
		scheduleAfter = scheduler.TrySchedule()
		// 重置调度间隔
		scheduleTimer.Reset(scheduleAfter)
	}
}

// PushJobEvent 推送任务变化事件
func (scheduler *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	scheduler.jobEventChan <- jobEvent
}

func InitScheduler() (err error) {
	GScheduler = &Scheduler{
		jobEventChan: make(chan *common.JobEvent, 1000), // 1000的容量
		jobPlanTable: make(map[string]*common.JobSchedulePlan),
		jobExecutingTable: make(map[string]*common.JobExecuteInfo),
		jobResultChan: make(chan *common.JobExecuteResult, 1000),
	}

	go GScheduler.schedulerLoop()
	return
}

func (scheduler *Scheduler) PushJobResult(result *common.JobExecuteResult)  {
	GScheduler.jobResultChan <- result
}
