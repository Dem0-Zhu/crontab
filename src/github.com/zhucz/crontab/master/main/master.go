package main

import (
	"flag"
	"fmt"
	"runtime"
	"task_scheduler/src/github.com/zhucz/crontab/master"
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
	// 启动命令 master -config ./master.json
	flag.StringVar(&confFile, "config", "./master.json", "指定master.json")
	flag.Parse()
}

func main() {
	var (
		err error
	)

	initArgs()
	initEnv()

	//加载配置
	if err = master.InitConfig(confFile); err != nil {
		goto ERR
	}

	// 日志管理
	if err = master.InitLogMgr(); err != nil {
		goto ERR
	}

	if err = master.InitJobMgr(); err != nil {
		goto ERR
	}

	// 启动API HTTP服务
	if err = master.InitApiServer(); err != nil {
		goto ERR
	}

	select {}

	return

ERR:
	fmt.Println(err)
}
