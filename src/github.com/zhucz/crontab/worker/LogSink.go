package worker

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"task_scheduler/src/github.com/zhucz/crontab/common"
	"time"
)

// mongodb存储日志
type LogSink struct {
	client     *mongo.Client
	logCollection *mongo.Collection
	logChan chan *common.JobLog
}

var (
	G_logSink *LogSink
)

// 批量写入日志
func (logSink *LogSink) saveLogs(batch *common.LogBatch)  {
	if _, err := logSink.logCollection.InsertMany(context.TODO(), batch.Logs); err != nil {
		fmt.Println("日志写入失败", err)
	}
}

// 日志存储协程
func (logSink *LogSink) writeLoop() {
	var (
		// 按批次写入mongo
		logs *common.LogBatch
	)
	for {
		select {
		case log := <- G_logSink.logChan:
			if logs == nil {
				logs = &common.LogBatch{}
			}
			logs.Logs = append(logs.Logs, log)
			if len(logs.Logs) >= G_config.JobLogBatchSize {
				// todo: 另起一个协程，注意安全问题
				logSink.saveLogs(logs)
				logs = nil
			}
		}
	}
}
func InitLogSink() (err error) {
	var (
		client *mongo.Client
		clientOption *options.ClientOptions
		conTimeout time.Duration
	)
	conTimeout = time.Duration(G_config.MongodbConnectTimeout) * time.Millisecond)
	clientOption = &options.ClientOptions{
		Hosts: []string{G_config.MongodbUri},
		ConnectTimeout: &conTimeout,
	}
	if client, err = mongo.Connect(context.TODO(), clientOption); err != nil {
		return
	}

	G_logSink = &LogSink{
		client: client,
		logCollection: client.Database("cron").Collection("log"),
		logChan: make(chan *common.JobLog, 1000),
	}

	// 启动mongodb处理协程
	go G_logSink.writeLoop()
	return
}
