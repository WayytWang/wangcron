package main

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

type result struct {
	err    error
	output []byte
}

func main() {
	//执行1个cmd,让它在一个协程里去执行,让它执行2秒
	//1秒的时候,杀死cmd

	var (
		ctx        context.Context
		cancelFunc context.CancelFunc
		cmd        *exec.Cmd
		res        *result
		resultChan chan *result
	)

	//context: chan byte
	//cancelFunc close(chan byte)

	resultChan = make(chan *result, 1000)

	ctx, cancelFunc = context.WithCancel(context.TODO())

	go func() {
		var (
			output []byte
			err    error
		)
		cmd = exec.CommandContext(ctx, "C:\\cygwin64\\bin\\bash.exe", "-c", "/usr/bin/sleep 2;/usr/bin/echo hello")

		output, err = cmd.CombinedOutput()

		resultChan <- &result{
			output: output,
			err:    err,
		}

	}()

	//继续往下走
	time.Sleep(1 * time.Second)
	cancelFunc()

	//在main协程里,等待子协程的退出,并打印任务执行结果
	res = <-resultChan

	fmt.Println(res.err, string(res.output))
}
