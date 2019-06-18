package master

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"time"
	"wangcron/crontab/common"
)

//APIServer 任务的HTTP接口
type APIServer struct {
	httpServer *http.Server
}

var (
	//G_apiServer 单例对象
	G_apiServer *APIServer
)

//保存任务接口
func handleJobSave(w http.ResponseWriter, r *http.Request) {
	//任务保存到ETCD中
	var (
		err     error
		postJob string
		job     common.Job
		oldJob  *common.Job
		bytes   []byte
	)

	// 1, 解析POST表单
	if err = r.ParseForm(); err != nil {
		goto ERR
	}

	// 2, 取表单中的job字段
	postJob = r.PostForm.Get("job")

	//3.反序列化job
	if err = json.Unmarshal([]byte(postJob), &job); err != nil {
		goto ERR
	}

	//4.保存到etcd
	if oldJob, err = G_jobMgr.SaveJob(&job); err != nil {
		goto ERR
	}

	//5.返回正常应答
	if bytes, err = common.BuildResponse(0, "success", oldJob); err == nil {
		w.Write(bytes)
	}

	return

ERR:
	//6.返回异常应答
	if bytes, err = common.BuildResponse(-1, err.Error(), oldJob); err == nil {
		w.Write(bytes)
	}
}

//删除任务接口
func handleJobDelete(w http.ResponseWriter, r *http.Request) {
	var (
		err    error
		name   string
		bytes  []byte
		oldJob *common.Job
	)

	if err = r.ParseForm(); err != nil {
		goto ERR
	}

	//删除的任务名
	name = r.PostForm.Get("name")

	//ectd中删除任务
	if oldJob, err = G_jobMgr.DeleteJob(name); err != nil {
		goto ERR
	}

	//正常应答
	if bytes, err = common.BuildResponse(0, "success", oldJob); err == nil {
		w.Write(bytes)
	}

	return

ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		w.Write(bytes)
	}
}

//任务列表
func handleJobList(w http.ResponseWriter, r *http.Request) {
	var (
		jobList []*common.Job
		bytes   []byte
		err     error
	)

	//获取任务列表
	if jobList, err = G_jobMgr.ListJobs(); err != nil {
		goto ERR
	}

	//正常应答
	if bytes, err = common.BuildResponse(0, "success", jobList); err == nil {
		w.Write(bytes)
	}
	return

ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		w.Write(bytes)
	}
}

//强制杀死任务
func handleJobKill(w http.ResponseWriter, r *http.Request) {
	var (
		err   error
		name  string
		bytes []byte
	)

	if err = r.ParseForm(); err != nil {
		goto ERR
	}

	//要杀死的任务名
	name = r.PostForm.Get("name")

	//杀死任务
	if err = G_jobMgr.KillJob(name); err != nil {
		goto ERR
	}

	//正常应答
	if bytes, err = common.BuildResponse(0, "success", nil); err == nil {
		w.Write(bytes)
	}
	return

ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		w.Write(bytes)
	}
}

//查询任务日志
func handleJobLog(w http.ResponseWriter, r *http.Request) {
	var (
		err        error
		bytes      []byte
		name       string //任务名字
		skipParam  string //从第几条开始
		skip       int
		limitParam string //返回多少条
		limit      int
		logArr     []*common.JobLog
	)

	//解析GET参数
	if err = r.ParseForm(); err != nil {
		goto ERR
	}

	//获取请求参数 /job/log?name=job10&skip=0&limit=10
	name = r.Form.Get("name")
	skipParam = r.Form.Get("skip")
	limitParam = r.Form.Get("limit")

	if skip, err = strconv.Atoi(skipParam); err != nil {
		skip = 0
	}

	if limit, err = strconv.Atoi(limitParam); err != nil {
		limit = 20
	}

	if logArr, err = G_logMgr.ListLog(name, skip, limit); err != nil {
		goto ERR
	}

	//正常应答
	if bytes, err = common.BuildResponse(0, "success", logArr); err == nil {
		w.Write(bytes)
	}
	return

ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		w.Write(bytes)
	}
}

//查看健康节点列表
func handleWorkerList(w http.ResponseWriter, r *http.Request) {
	var (
		workerArr []string
		err       error
		bytes     []byte
	)

	if workerArr, err = G_workerMgr.ListWorkers(); err != nil {
		goto ERR
	}

	// 正常应答
	if bytes, err = common.BuildResponse(0, "success", workerArr); err == nil {
		w.Write(bytes)
	}
	return

ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		w.Write(bytes)
	}
}

//InitAPIServer 初始化服务
func InitAPIServer() (err error) {
	var (
		mux           *http.ServeMux
		listener      net.Listener
		httpServer    *http.Server
		staticDir     http.Dir
		staticHandler http.Handler
	)

	//配置路由
	mux = http.NewServeMux()
	mux.HandleFunc("/job/save", handleJobSave)
	mux.HandleFunc("/job/delete", handleJobDelete)
	mux.HandleFunc("/job/list", handleJobList)
	mux.HandleFunc("/job/kill", handleJobKill)
	mux.HandleFunc("/job/log", handleJobLog)
	mux.HandleFunc("/worker/list", handleWorkerList)

	//静态文件目录
	staticDir = http.Dir(G_config.WebRoot)
	staticHandler = http.FileServer(staticDir)
	mux.Handle("/", http.StripPrefix("/", staticHandler))

	//启动TCP监听
	if listener, err = net.Listen("tcp", ":"+strconv.Itoa(G_config.ApiPort)); err != nil {
		return
	}

	//创建一个HTTP服务
	httpServer = &http.Server{
		ReadTimeout:  time.Duration(G_config.ApiReadTimeOut) * time.Millisecond,
		WriteTimeout: time.Duration(G_config.ApiWriteTimeOut) * time.Millisecond,
		Handler:      mux,
	}

	G_apiServer = &APIServer{
		httpServer: httpServer,
	}

	go httpServer.Serve(listener)

	return
}
