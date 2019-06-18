package master

import (
	"context"
	"time"
	"wangcron/crontab/common"

	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
)

//LogMgr 日志管理
type LogMgr struct {
	client        *mongo.Client
	logCollection *mongo.Collection
}

var (
	G_logMgr *LogMgr
)

//InitLogMgr 初始化日志管理
func InitLogMgr() (err error) {
	var (
		client *mongo.Client
	)

	//建立mongodb连接
	if client, err = mongo.Connect(context.TODO(),
		G_config.MongodbURI,
		options.Client().SetConnectTimeout(time.Duration(G_config.MongodbConnectTimeout)*time.Second)); err != nil {
		return
	}

	G_logMgr = &LogMgr{
		client:        client,
		logCollection: client.Database("cron").Collection("log"),
	}
	return
}

//ListLog 查看任务日志
func (logMgr *LogMgr) ListLog(name string, skip, limit int) (logArr []*common.JobLog, err error) {
	var (
		filter  *common.JobLogFilter
		logSort *common.SortLogByStartTime
		cursor  mongo.Cursor
		jobLog  *common.JobLog
	)

	logArr = make([]*common.JobLog, 0)

	//过滤条件
	filter = &common.JobLogFilter{JobName: name}

	//按照任务开始时间倒排
	logSort = &common.SortLogByStartTime{SortOrder: -1}

	if cursor, err = logMgr.logCollection.Find(context.TODO(),
		filter,
		options.Find().SetSort(logSort),
		options.Find().SetSkip(int64(skip)),
		options.Find().SetLimit(int64(limit))); err != nil {
		return
	}

	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		jobLog = &common.JobLog{}

		//反序列化bson
		if err = cursor.Decode(jobLog); err != nil {
			continue //有日志不合法
		}

		logArr = append(logArr, jobLog)
	}
	return
}
