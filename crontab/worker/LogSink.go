package worker

import (
	"context"
	"time"
	"wangcron/crontab/common"

	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
)

//LogSink mongodb存储日志
type LogSink struct {
	client         *mongo.Client
	logCollection  *mongo.Collection
	logChan        chan *common.JobLog
	autoCommitChan chan *common.LogBatch
}

var (
	G_logSink *LogSink
)

//发送日志到mongodb
func (logSink *LogSink) saveLogs(bacth *common.LogBatch) {
	logSink.logCollection.InsertMany(context.TODO(), bacth.Logs)
}

//日志存储协程
func (logSink *LogSink) writeLoop() {
	var (
		log          *common.JobLog
		logBatch     *common.LogBatch
		commitTimer  *time.Timer
		timeOutBatch *common.LogBatch
	)

	for {
		select {
		case log = <-logSink.logChan:
			//把这条log写到mongodb中
			//logSink.lokCollection.InsertOne
			//每次插入需要等待mongodb的一次请求往返,耗时可能因为网络慢花费比较长的时间

			if logBatch == nil {
				logBatch = &common.LogBatch{}

				//让这个批次超时自动提交(给1秒时间)
				commitTimer = time.AfterFunc(time.Duration(G_config.JobLogCommitTimeout)*time.Millisecond, func(batch *common.LogBatch) func() {
					return func() {
						logSink.autoCommitChan <- batch
					}
				}(logBatch))
			}

			//把新日志追加到批次中
			logBatch.Logs = append(logBatch.Logs, log)

			//如果批次满了,就立即发送
			if len(logBatch.Logs) >= G_config.JobLobBatchSize {
				//发送日志
				logSink.saveLogs(logBatch)
				logBatch = nil
				//取消定时器
				commitTimer.Stop()
			}
		case timeOutBatch = <-logSink.autoCommitChan: //过期的批次
			//判断过期批次是否仍旧是当前的批次
			if timeOutBatch != logBatch {
				continue //跳过
			}

			//把批次写入mongodb中
			logSink.saveLogs(timeOutBatch)

			//清空logBatch
			logBatch = nil
		}
	}
}

//Append 发送日志
func (logSink *LogSink) Append(jobLog *common.JobLog) {
	select {
	case logSink.logChan <- jobLog:
	default:
		//队列满了就丢弃
	}
}

//InitLogSink 初始化日志打印
func InitLogSink() (err error) {
	var (
		client *mongo.Client
	)

	//建立mongodb连接
	if client, err = mongo.Connect(context.TODO(),
		G_config.MongodbURI,
		options.Client().SetConnectTimeout(time.Duration(G_config.MongodbConnectTimeout)*time.Second)); err != nil {
		return
	}

	//选择db和collection
	G_logSink = &LogSink{
		client:         client,
		logCollection:  client.Database("cron").Collection("log"),
		logChan:        make(chan *common.JobLog, 1000),
		autoCommitChan: make(chan *common.LogBatch, 1000),
	}

	//启动处理协程
	go G_logSink.writeLoop()

	return
}
