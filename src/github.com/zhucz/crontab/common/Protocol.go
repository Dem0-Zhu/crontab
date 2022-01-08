package common

import (
	"encoding/json"
)

type Job struct {
	Name     string `json:"name"`      // 任务名
	Command  string `json:"command"`   // shell命令
	CronExpr string `json:"cron_expr"` // cron表达式
}

// Response HTTP接口应答
type Response struct {
	Errno int         `json:"errno"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}

// BuildResponse 应答方法
func BuildResponse(errno int, msg string, data interface{}) (resp []byte, err error) {
	response := Response{
		Errno: errno,
		Msg:   msg,
		Data:  data,
	}
	resp, err = json.Marshal(response)
	return
}
