// Copyright (c) 2013 Beijing CmsTop Technology Co.,Ltd. (http://www.cmstop.com)

/*
Package sfss provides SFSS Daemon.
SFSS provides site & db management api.
SFSS provides site & db usage Report.
protocol details in wiki http://wiki.9466.dev/
*/
package main

import (
	"fmt"
	"github.com/9466/goconfig"
	"log"
	"os"
	"os/signal"
	"sfss/server"
	"sfss/util"
	"syscall"
	"time"
)

const (
	PROCESS_NUM = 1 // 系统当前会启动的进程数，目前只有2个，分别是server和monitor
)

func main() {
	// 初始化配置文件
	var err error
	var dir string
	dir, err = util.GetDir()
	if err != nil {
		log.Fatalln("GetDir Error:", err.Error())
	}
	var configFile string = dir + "/conf/sfss.conf"
	sfss := new(util.SFSS)
	sfss.ConnNum = 0
	sfss.Conf, err = goconfig.ReadConfigFile(configFile)
	if err != nil {
		log.Fatalln("ReadConfigFile Err: ", err.Error(), "\nConfigFile:", configFile)
	}

	// 初始化日志
	logFile, err := sfss.Conf.GetString("log", "file")
	if err != nil {
		log.Fatalln("ConfigFile Parse Error.", err.Error())
	}
	if logFile[0] != '/' {
		logFile = dir + "/" + logFile
	}
	logFileHandle, err := os.OpenFile(logFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	defer logFileHandle.Close()
	if err != nil {
		log.Fatalln(err.Error())
	}
	sfss.Logger = log.New(logFileHandle, "", log.Ldate|log.Ltime|log.Lshortfile)

	// 开始启动服务
	sfss.Logger.Println("SFSS starting...")
	sfss.Chs = make(chan int, PROCESS_NUM) // 初始化channel数量
	sfss.Shutdown = false                  // 默认关闭状态为false

	// 启动一个新的server
	sfssSever, err := server.NewServer(sfss)
	if err != nil {
		sfss.Logger.Fatalln(err.Error())
	}

	// 开始服务
	go sfssSever.Accept()

	// 启动数据上报服务

	// 开始服务

	// 监听系统信号，重启或停止服务
	// trap signal
	sch := make(chan os.Signal, 10)
	signal.Notify(sch, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT,
		syscall.SIGHUP, syscall.SIGSTOP, syscall.SIGQUIT)
	go func(ch <-chan os.Signal) {
		sig := <-ch
		sfss.Logger.Println("signal recieved " + sig.String() + ", at: " + time.Now().String())
		sfss.Shutdown = true
		sfssSever.Close()
		if sig == syscall.SIGHUP {
			sfss.Logger.Println("SFSS restart now...")
			procAttr := new(os.ProcAttr)
			procAttr.Files = []*os.File{nil, os.Stdout, os.Stderr}
			procAttr.Dir = os.Getenv("PWD")
			procAttr.Env = os.Environ()
			process, err := os.StartProcess(os.Args[0], os.Args, procAttr)
			if err != nil {
				sfss.Logger.Println("SFSS restart process failed:" + err.Error())
				return
			}
			waitMsg, err := process.Wait()
			if err != nil {
				sfss.Logger.Println("SFSS restart wait error:" + err.Error())
			}
			sfss.Logger.Println(waitMsg)
		} else {
			sfss.Logger.Println("SFSS shutdown now...")
		}
	}(sch)

	// DEBUG模式处理
	DEBUG, _ := sfss.Conf.GetBool("server", "debug")
	if DEBUG {
		// DEBUG: 输出连接数
		go func(s *util.SFSS) {
			tick := time.Tick(time.Second)
			for now := range tick {
				fmt.Printf("%v SFSS.ConnNum %d\n", now, s.ConnNum)
			}
		}(sfss)

		// DEBUG: 关闭测试
		debugTime, _ := sfss.Conf.GetInt64("server", "debugTime")
		if debugTime > 0 {
			go func(s *util.SFSS) {
				time.Sleep(time.Duration(debugTime) * time.Second)
				fmt.Println("timed out")
				s.Shutdown = true
				sfssSever.Close()
			}(sfss)
		}
	}

	// 启动chs监听，等待所有Process都运行结束
	for i := 0; i < PROCESS_NUM; i++ {
		<-sfss.Chs
	}
	sfss.Logger.Println("SFSS stopped.")
}
