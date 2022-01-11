package master

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"task_scheduler/src/github.com/zhucz/crontab/common"
	"time"
)

type LogMgr struct {
	client     *mongo.Client
	logCollection *mongo.Collection
}

var (
	G_logMgr *LogMgr
)

func InitLogMgr() (err error) {
	var (
		client *mongo.Client
		conTimeout time.Duration
		clientOption *options.ClientOptions
	)

	// 建立mongodb连接
	conTimeout = time.Duration(G_config.MongodbConnectTimeout) * time.Millisecond
	clientOption = &options.ClientOptions{
		Hosts: []string{G_config.MongodbUri},
		ConnectTimeout: &conTimeout,
	}
	if client, err = mongo.Connect(context.TODO(), clientOption); err != nil {
		return
	}

	G_logMgr = &LogMgr{
		client: client,
		logCollection: client.Database("cron").Collection("log"),
	}
	return
}

func (logMgr *LogMgr) ListLog(name string, skip int64, limit int64) (logArr []*common.JobLog, err error) {
	var (
		filter *common.JobLogFilter
		logSort *common.SortLogByStartTime
		cursor *mongo.Cursor
		jobLog *common.JobLog
	)
	logArr = make([]*common.JobLog, 0)
	filter = &common.JobLogFilter{
		JobName: name,
	}
	// 按开始时间倒排
	logSort = &common.SortLogByStartTime{StartTime: -1}
	if cursor, err = G_logMgr.logCollection.Find(context.TODO(), filter, &options.FindOptions{Sort: logSort, Limit: &limit, Skip: &skip}); err != nil {
		return
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		jobLog = &common.JobLog{}
		// 反序列化

		if err = cursor.Decode(jobLog); err != nil {
			continue
		}
		logArr = append(logArr, jobLog)
	}
	return
}
