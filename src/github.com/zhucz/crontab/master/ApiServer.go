package master

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"task_scheduler/src/github.com/zhucz/crontab/common"
	"time"
)

// G_apiServer 单例对象
var G_apiServer *ApiServer

// ApiServer 任务的HTTp接口
type ApiServer struct {
	httpServer *http.Server
}

// 保存任务接口
// POST job={"name":"job1", "command":"echo hello", "cronExpr":"* * * * *"}
func handleJobPut(resp http.ResponseWriter, req *http.Request) {
	// 任务保存到ETCD中
	var (
		err     error
		postJob string
		job     common.Job
		oldJob  *common.Job
		result  []byte
	)
	//解析post表单
	if err = req.ParseForm(); err != nil {
		goto ERR
	}
	// 取表单中的job字段
	postJob = req.PostForm.Get("job")
	// 反序列化Job
	if err = json.Unmarshal([]byte(postJob), &job); err != nil {
		goto ERR
	}
	// 保存到etcd
	if oldJob, err = G_jobMgr.PutJob(&job); err != nil {
		goto ERR
	}
	// 返回正常应答
	// 默认BuildResponse不会返回错误
	if result, err = common.BuildResponse(0, "success", oldJob); err == nil {
		resp.Write(result)
	}
	return
ERR:
	if result, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(result)
	}
}

// POST /job/get/ name=job1
func handleJobGet(resp http.ResponseWriter, req *http.Request)  {
	var (
		err error
		name string
		job *common.Job
		result []byte
	)
	if err = req.ParseForm(); err != nil {
		goto ERR
	}
	name = req.PostForm.Get("name")
	if job, err = G_jobMgr.GetJob(name); err != nil {
		goto ERR
	}
	if result, err = common.BuildResponse(0, "success", job); err == nil {
		resp.Write(result)
	}
	return
ERR:
	if result, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(result)
	}
}

// POST /job/delete/ name=job1
func handleJobDelete(resp http.ResponseWriter, req *http.Request) {
	var (
		err error
		name string
		oldJob *common.Job
		result []byte
	)
	if err = req.ParseForm(); err != nil {
		goto ERR
	}
	name = req.PostForm.Get("name")
	if oldJob, err = G_jobMgr.DeleteJob(name); err != nil {
		goto ERR
	}
	if result, err = common.BuildResponse(0, "success", oldJob); err == nil {
		resp.Write(result)
	}
	return
ERR:
	if result, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(result)
	}
}

func handleJobList(resp http.ResponseWriter, req *http.Request) {
	var (
		jobList map[string]common.Job
		err error
		result []byte
	)
	if jobList, err = G_jobMgr.ListJobs(); err != nil {
		goto ERR
	}

	if result, err = common.BuildResponse(0, "success", jobList); err == nil {
		resp.Write(result)
	}
	return

ERR:
	if result, err = common.BuildResponse(0, err.Error(), nil); err == nil {
		resp.Write(result)
	}
}

// POST /job/kill name=job1
func handleJobKill(resp http.ResponseWriter, req *http.Request)  {
	var (
		name string
		err error
		result []byte
	)
	if err = req.ParseForm(); err != nil {
		goto ERR
	}
	name = req.PostForm.Get("name")
	fmt.Println("name ", name)
	if err = G_jobMgr.KillJob(name); err != nil {
		goto ERR
	}
	if result, err = common.BuildResponse(0, "success", nil); err == nil {
		resp.Write(result)
	}
	return
ERR:
	if result, err = common.BuildResponse(-1, "fail", nil); err == nil {
		resp.Write(result)
	}
}
// GET job/log?name=xxx&skip=1&limit=1
func handleJobLog(resp http.ResponseWriter, req *http.Request) {
	var (
		err error
		result []byte
		name string
		skip int
		limit int
		logArr []*common.JobLog
	)
	if err = req.ParseForm(); err != nil {
		goto ERR
	}
	name = req.Form.Get("name")
	if skip, err  = strconv.Atoi(req.Form.Get("skip")); err != nil {
		skip = 0
	}
	if limit, err = strconv.Atoi(req.Form.Get("limit")); err != nil {
		limit = 10
	}
	if logArr, err = G_logMgr.ListLog(name, int64(skip), int64(limit)); err != nil {
		goto ERR
	}
	if result, err = common.BuildResponse(0, "success", logArr); err == nil {
		resp.Write(result)
	}
	return

ERR:
	if result, err = common.BuildResponse(-1, "fail", nil); err == nil {
		resp.Write(result)
	}
}

// 获取健康worker节点列表
func handleWorkerList(resp http.ResponseWriter, res *http.Request)  {
	var (
		workerArr []string
		err error
		result []byte
	)
	if workerArr, err = G_workerMgr.ListWorkers(); err != nil {
		goto ERR
	}
	if result, err = common.BuildResponse(0, "success", workerArr); err == nil {
		resp.Write(result)
	}
	return
ERR:
	if result, err = common.BuildResponse(-1, "fail", nil); err == nil {
		resp.Write(result)
	}
}

// InitApiServer 初始化服务
func InitApiServer() (err error) {
	// 配置路由
	var (
		mux *http.ServeMux
		listen net.Listener
		server *http.Server
		staticDir http.Dir
		staticHandle http.Handler
	)
	mux = http.NewServeMux()
	mux.HandleFunc("/job/put", handleJobPut)
	mux.HandleFunc("/job/get", handleJobGet)
	mux.HandleFunc("/job/delete", handleJobDelete)
	mux.HandleFunc("/job/list", handleJobList)
	mux.HandleFunc("/job/kill", handleJobKill)
	mux.HandleFunc("/job/log", handleJobLog)
	mux.HandleFunc("/worker/list", handleWorkerList)

	staticDir = http.Dir(G_config.WebRoot)
	staticHandle = http.FileServer(staticDir)
	mux.Handle("/", http.StripPrefix("/", staticHandle))

	// 启动TCP监听
	if listen, err = net.Listen("tcp", ":"+strconv.Itoa(G_config.ApiPort)); err != nil {
		return
	}

	// 创建一个HTTP服务
	/*
		当http.Server收到请求之后，会回调给Handler方法（mux），mux会根据请求的url，找到匹配的回调函数，转发。
	*/
	server = &http.Server{
		ReadTimeout:  time.Duration(G_config.ApiReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(G_config.ApiWriteTimeout) * time.Millisecond,
		Handler:      mux,
	}

	G_apiServer = &ApiServer{
		httpServer: server,
	}

	// 启动服务端
	go server.Serve(listen)

	return
}
