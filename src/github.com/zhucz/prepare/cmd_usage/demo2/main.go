package main

import (
	"fmt"
	"os/exec"
)

func main() {
	var (
		cmd    *exec.Cmd
		output []byte
		err    error
	)
	// 生成cmd
	cmd = exec.Command("C:\\Windows\\System32\\bash.exe", "-c", "sleep 5;ls -l;echo hello")
	// 执行命令并捕获子进程输出（pipe）
	if output, err = cmd.CombinedOutput(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(output))
}
