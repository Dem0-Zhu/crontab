package main

import (
	"flag"
	"fmt"
	"runtime"
	"task_scheduler/src/github.com/zhucz/crontab/worker"
)

var (
	confFile string // 配置文件路径
)

// 初始化线程
func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

// 解析命令行参数
func initArgs()  {
	// 启动命令 worker -config ./worker.json
	flag.StringVar(&confFile, "config", "./worker.json", "指定worker.json")
	flag.Parse()
}

func main() {
	var (
		err error
	)

	initArgs()
	initEnv()

	//加载配置
	if err = worker.InitConfig(confFile); err != nil {
		goto ERR
	}

	if err = worker.InitJobMgr(); err != nil {
		goto ERR
	}


	for {}

	return

ERR:
	fmt.Println(err)
}
