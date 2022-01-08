package main

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

type Result struct {
	err    error
	output []byte
}

// 杀死command
func main() {
	var (
		cmd        *exec.Cmd
		output     []byte
		err        error
		res        *Result
		resultChan chan *Result
	)
	resultChan = make(chan *Result, 100)
	// 一个协程去执行一个command，让它执行两秒；在一秒的时候，杀死它；
	cxt, cancelFunc := context.WithCancel(context.TODO())
	go func() {
		cmd = exec.CommandContext(cxt, "C:\\Windows\\System32\\bash.exe", "-c", "sleep 2;echo hello;")
		output, err = cmd.CombinedOutput()
		resultChan <- &Result{
			err:    err,
			output: output,
		}

	}()

	time.Sleep(1 * time.Second)

	// 取消上下文
	cancelFunc()

	res = <-resultChan
	fmt.Println(res.err, string(res.output))
}
