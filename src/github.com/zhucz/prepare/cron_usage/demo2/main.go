package main

import (
	"fmt"
	"time"
)

import (
	"github.com/gorhill/cronexpr"
)

/*
	实现简易的任务调度
*/

type CronJob struct {
	expr     *cronexpr.Expression
	nextTime time.Time // 下次应该被调度的时间
}

func main() {
	// 需要一个调度协程，定时检查所有cron任务，发现过期的就执行
	var (
		expr          *cronexpr.Expression
		cronJob       *CronJob
		scheduleTable map[string]*CronJob
	)
	scheduleTable = make(map[string]*CronJob)

	now := time.Now()

	expr = cronexpr.MustParse("*/5 * * * * * *")
	cronJob = &CronJob{
		expr:     expr,
		nextTime: expr.Next(now),
	}
	scheduleTable["Job1"] = cronJob
	expr = cronexpr.MustParse("*/5 * * * * * *")
	cronJob = &CronJob{
		expr:     expr,
		nextTime: expr.Next(now),
	}
	scheduleTable["Job2"] = cronJob

	// 启动一个调度协程
	go func() {
		for {
			now := time.Now()
			for name, cronJob := range scheduleTable {
				if cronJob.nextTime.Before(now) || cronJob.nextTime.Equal(now) {
					go func(name string) {
						fmt.Printf("执行 jobName = %s\n", name)
					}(name)
					// 计算下一次调度时间
					cronJob.nextTime = expr.Next(now)
					fmt.Println("下次调度时间为 ： ", cronJob.nextTime)
				}
			}
			select {
			case <-time.NewTimer(100).C:
			}
		}

	}()

	time.Sleep(500 * time.Second)

}
