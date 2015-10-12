// Copyright (c) 2013 Beijing CmsTop Technology Co.,Ltd. (http://www.cmstop.com)

package util

import (
	"github.com/9466/goconfig"
	"log"
)

// 系统公共数据结构
type SFSS struct {
	Conf     *goconfig.ConfigFile // 配置文件接口
	Logger   *log.Logger          // 日志处理接口
	Chs      chan int             // 进程处理channel
	Shutdown bool                 // 程序是否关闭标识
	ConnNum  int                  // 当前TCP连接数
}

// 接收数据结构
type ReceiveData struct {
	Serverid int    `json:"serverid"` // 服务器编号
	Data     string `json:"data"`     // 加密后的数据
}

// 响应数据结构
type SendData struct {
	Code    int    `json:"code"`    // 状态码，0 表示成功，非0表示失败
	Message string `json:"message"` // 消息字符串
}

// 业务数据结构
type OrderData struct {
	Method string            `json:"method"` // 业务操作类型
	Data   map[string]string `json:"data"`   // 业务操作数据
}

// 测试接口返回数据结构
type InitTestData struct {
	SendData
	Data map[string]interface{} `json:"data"` // 数据
}
