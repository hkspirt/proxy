//----------------
//Func  :
//Author: xjh
//Date  : 2019/
//Note  :
//----------------
package util

import (
	"crypto/md5"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync/atomic"
	"syscall"
)

func Md5(str string) string {
	hash := md5.New()
	hash.Write([]byte(str))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func Try(fun func(), handler func(interface{})) {
	defer func() {
		if err := recover(); err != nil {
			if handler == nil {
				LogError("panic catch:%v stack:%s", err, CallFuncStack())
			} else {
				handler(err)
			}
		}
	}()
	fun()
}

var (
	stopFromSys = make(chan os.Signal, 1)
	stopSignal  = make(chan bool)
	isRunning   = int32(1)
)

func Stop() {
	if atomic.CompareAndSwapInt32(&isRunning, 1, 0) {
		close(stopSignal)
	}
}

func WaitForSystemExit() {
	signal.Notify(stopFromSys, os.Interrupt, os.Kill, syscall.SIGTERM)
	select {
	case <-stopFromSys:
		Stop()
	}
}

func IsRuning() bool {
	return atomic.LoadInt32(&isRunning) == 1
}

func GoChan(fn func(cstop <-chan bool)) {
	go func() {
		Try(func() {
			fn(stopSignal)
		}, nil)
	}()
}

func Go(fn func()) {
	go func() {
		Try(fn, nil)
	}()
}

func CallFuncStack() string {
	buf := make([]byte, 1<<12)
	return string(buf[:runtime.Stack(buf, false)])
}
