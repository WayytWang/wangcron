package worker

import (
	"fmt"
	"time"
	"wangcron/crontab/common"
)

//Scheduler 任务调度器
type Scheduler struct {
	jobEventChan      chan *common.JobEvent              //etcd任务事件队列
	jobPlanTable      map[string]*common.JobSchedulePlan //任务调度计划表
	jobExecutingTable map[string]*common.JobExecuteInfo  //任务执行计划表
	jobResultChan     chan *common.JobExecuteResult      //任务执行结果队列
}

var (
	G_scheduler *Scheduler
)

//处理任务事件
func (scheduler *Scheduler) handleJobEvent(jobEvent *common.JobEvent) {
	var (
		jobchedulePlan *common.JobSchedulePlan
		jobExecuteInfo *common.JobExecuteInfo
		jobExecuting   bool
		jobExisted     bool
		err            error
	)

	switch jobEvent.EventType {
	case common.JOB_EVENT_SAVE: //保存任务事件
		if jobchedulePlan, err = common.BuildJobSchedulePlan(jobEvent.Job); err != nil {
			return
		}
		scheduler.jobPlanTable[jobEvent.Job.Name] = jobchedulePlan
	case common.JOB_EVENT_DELETE: //删除任务事件
		if jobchedulePlan, jobExisted = scheduler.jobPlanTable[jobEvent.Job.Name]; jobExisted {
			delete(scheduler.jobPlanTable, jobEvent.Job.Name)
		}
	case common.JOB_EVENT_KILL: //强杀任务事件
		//取消掉Command执行,判断任务是否在执行中
		if jobExecuteInfo, jobExecuting = scheduler.jobExecutingTable[jobEvent.Job.Name]; jobExecuting {
			jobExecuteInfo.CancelFunc() //触发command杀死shell指令 任务退出
		}

	}
}

//处理任务结果
func (scheduler *Scheduler) handleJobResult(result *common.JobExecuteResult) {
	var (
		jobLog *common.JobLog
	)

	//删除执行状态
	delete(scheduler.jobExecutingTable, result.ExecuteInfo.Job.Name)

	//生成执行日志
	if result.Err != common.ERR_LOCK_ALREADY_REQUIRED {
		jobLog = &common.JobLog{
			JobName:      result.ExecuteInfo.Job.Name,
			Command:      result.ExecuteInfo.Job.Command,
			Output:       string(result.Output),
			PlanTime:     result.ExecuteInfo.PlanTime.UnixNano() / 1000 / 1000,
			ScheduleTime: result.ExecuteInfo.RealTime.UnixNano() / 1000 / 1000,
			StartTime:    result.StartTime.UnixNano() / 1000 / 1000,
			EndTime:      result.EndTime.UnixNano() / 1000 / 1000,
		}

		if result.Err != nil {
			jobLog.Err = result.Err.Error()
		} else {
			jobLog.Err = ""
		}

		//存储到mongodb
		G_logSink.Append(jobLog)
	}

	fmt.Println("任务执行完成:", result.ExecuteInfo.Job.Name, string(result.Output), result.Err)
}

//TryStartJob 尝试执行任务
func (scheduler *Scheduler) TryStartJob(jobPlan *common.JobSchedulePlan) {
	//调度 和 执行 是2件事情
	var (
		jobExecuteInfo *common.JobExecuteInfo
		jobExisted     bool
	)

	//执行任务可能运行很久(1分钟),1分钟可以调度60次,但是只能执行1次 还需要防止并发

	//如果任务正在执行,跳过本次调度
	if jobExecuteInfo, jobExisted = scheduler.jobExecutingTable[jobPlan.Job.Name]; jobExisted {
		fmt.Println("尚未退出,跳过执行:", jobExecuteInfo.Job.Name)
		return
	}

	//构建执行状态信息
	jobExecuteInfo = common.BuildJonExecuteInfo(jobPlan)

	//保存执行状态
	scheduler.jobExecutingTable[jobPlan.Job.Name] = jobExecuteInfo

	//执行任务
	G_executor.ExecuteJob(jobExecuteInfo)

}

//TrySchedule 重新计算任务调度状态
func (scheduler *Scheduler) TrySchedule() (schedulerAfter time.Duration) {
	//1.遍历所有任务
	//2.过期的任务立即执行
	//3.统计最近的要过期的任务的时间 (N秒后过期 == schedulerAfter)

	var (
		jobPlan  *common.JobSchedulePlan
		now      time.Time
		nearTime *time.Time
	)

	//如果任务表为空,随便给定一个睡眠时间
	if len(scheduler.jobPlanTable) == 0 {
		schedulerAfter = 1 * time.Second
		return
	}

	//当前时间
	now = time.Now()

	//遍历所有任务
	for _, jobPlan = range scheduler.jobPlanTable {
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {
			//TODO:尝试执行任务
			scheduler.TryStartJob(jobPlan)
			jobPlan.NextTime = jobPlan.Expr.Next(now) //更新下次执行时间
		}

		//统计最近一个要过期的任务
		if nearTime == nil || jobPlan.NextTime.Before(*nearTime) {
			nearTime = &jobPlan.NextTime
		}
	}

	//下次调度间隔 (最近要执行的任务调度时间 - 当前时间)
	schedulerAfter = (*nearTime).Sub(now)
	return
}

//调度协程
func (scheduler *Scheduler) schedulerLoop() {
	var (
		jobEvent       *common.JobEvent
		schedulerAfter time.Duration
		scheduleTimer  *time.Timer
		jobResult      *common.JobExecuteResult
	)

	//初始化一次
	schedulerAfter = scheduler.TrySchedule()

	//调度的延迟定时器
	scheduleTimer = time.NewTimer(schedulerAfter)

	//定时任务common.Job
	for {
		select {
		case jobEvent = <-scheduler.jobEventChan: //监听任务变化事件
			//对内存中维护的任务列表增删改查
			scheduler.handleJobEvent(jobEvent)
		case <-scheduleTimer.C: //最近任务到期了
		case jobResult = <-scheduler.jobResultChan: //监听任务执行结果
			scheduler.handleJobResult(jobResult)
		}

		//调度一次任务
		schedulerAfter = scheduler.TrySchedule()
		scheduleTimer.Reset(schedulerAfter)
	}
}

//PushJobEvent 推送任务变化事件
func (scheduler *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	scheduler.jobEventChan <- jobEvent
}

//InitScheduler 初始化任务调度器
func InitScheduler() (err error) {
	G_scheduler = &Scheduler{
		jobEventChan:      make(chan *common.JobEvent, 1000),
		jobPlanTable:      make(map[string]*common.JobSchedulePlan),
		jobExecutingTable: make(map[string]*common.JobExecuteInfo),
		jobResultChan:     make(chan *common.JobExecuteResult, 1000),
	}

	//启动调度协程
	go G_scheduler.schedulerLoop()
	return
}

//PushJobResult 回传任务执行结果
func (scheduler *Scheduler) PushJobResult(jobResult *common.JobExecuteResult) {
	scheduler.jobResultChan <- jobResult
}
