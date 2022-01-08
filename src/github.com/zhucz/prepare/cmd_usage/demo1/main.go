package main

import (
	"fmt"
	"os/exec"
)

func main() {
	var (
		cmd *exec.Cmd
		err error
	)

	cmd = exec.Command("C:\\Windows\\System32\\bash.exe", "-c", "echo hello")

	err = cmd.Run()

	fmt.Println(err)
}
