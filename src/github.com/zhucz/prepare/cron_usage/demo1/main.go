package main

import (
	"fmt"
	"github.com/gorhill/cronexpr"
	"time"
)

/*
	初识Cron
*/
func main() {
	var (
		expr     *cronexpr.Expression
		err      error
		now      time.Time
		nextTime time.Time
	)
	// 这个库比linux的crontab粒度更小，支持秒和年
	// 解析cron表达式 sec min hour day month week year
	//if expr, err = cronexpr.Parse("* * * * * * *"); err != nil {
	//	fmt.Println(err)
	//	return
	//}

	// 构造cron表达式，每五秒执行一次
	if expr, err = cronexpr.Parse("*/5 * * * * * *"); err != nil {
		fmt.Println(err)
		return
	}

	now = time.Now()
	// 计算下次执行的时间
	/*
		2021-11-03 22:47:52.5306221 +0800 CST m=+0.035022101 2021-11-03 22:50:00 +0800 CST
		这里需要注意：now=22:47  next = 22:50，原因是crontab执行时间为0、5、10、15、20...这种，并不是47+5；
	*/
	nextTime = expr.Next(now)
	// 定时器超时
	time.AfterFunc(nextTime.Sub(now), func() {
		fmt.Println("被调度了", nextTime)
	})

	fmt.Println(now, nextTime)

	time.Sleep(5 * time.Second)
	expr = expr
}
