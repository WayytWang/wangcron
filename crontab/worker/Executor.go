package worker

import (
	"fmt"
	"os/exec"
	"time"
	"wangcron/crontab/common"
)

//Executor 任务执行器
type Executor struct {
}

var (
	G_executor *Executor
)

//ExecuteJob 执行一个任务
func (executor *Executor) ExecuteJob(info *common.JobExecuteInfo) {
	fmt.Println("执行任务:", info.Job.Name, info.PlanTime, info.RealTime)

	go func() {
		var (
			cmd     *exec.Cmd
			err     error
			output  []byte
			result  *common.JobExecuteResult
			jobLock *JobLock
		)

		result = &common.JobExecuteResult{
			ExecuteInfo: info,
			Output:      make([]byte, 0),
		}

		//首先获取分布式锁
		//初始化分布式锁
		jobLock = G_jobMgr.CreateJobLock(info.Job.Name)

		//记录任务开始时间
		result.StartTime = time.Now()

		//抢锁
		err = jobLock.TryLock()
		defer jobLock.Unlock()

		if err != nil { //上锁失败
			result.Err = err
			result.EndTime = time.Now()
		} else {
			//抢锁成功后重置启动时间
			result.StartTime = time.Now()

			//执行shell命令
			cmd = exec.CommandContext(info.CancelCtx, "C:\\cygwin64\\bin\\bash.exe", "-c", info.Job.Command)

			//执行并捕获输出
			output, err = cmd.CombinedOutput()

			//记录任务结束时间
			result.EndTime = time.Now()
			result.Output = output
			result.Err = err
		}

		//任务执行完成后,把执行的结果返回给Scheduler,Scheduler会从executingTable中删除掉执行记录
		G_scheduler.PushJobResult(result)
	}()
}

//InitExecutor 初始化执行器
func InitExecutor() (err error) {
	G_executor = &Executor{}
	return
}
