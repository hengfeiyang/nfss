// Copyright (c) 2013 Beijing CmsTop Technology Co.,Ltd. (http://www.cmstop.com)

/*
Package server provides site & db management method.
It's a TCP Socket Server.
*/
package server

import (
	"encoding/json"
	"net"
	"sfss/util"
	"time"
)

const (
	DEF_SERVER_HOST = "0.0.0.0" // 默认服务地址
	DEF_SERVER_PORT = "9467"    // 默认服务端口号
)

// 服务器数据结构
type Serve struct {
	main       *util.SFSS       // 系统接口
	listen     *net.TCPListener // 服务监听接口
	host       string           // 服务地址
	port       string           // 服务端口
	serverIV   []byte           // 加密向量
	serverKEY  []byte           // 加密密钥
	serverType int              // 服务器服务类型
	site       *site            // 站点控制接口
	db         *db              // 数据库控制接口
}

// 创建一个新的服务器实例
func NewServer(s *util.SFSS) (*Serve, error) {
	server := new(Serve)
	server.main = s
	err := server.checkConfig()
	if err != nil {
		return nil, err
	}
	addr, err := net.ResolveTCPAddr("tcp4", server.host+":"+server.port)
	if err != nil {
		return nil, err
	}
	listener, err := net.ListenTCP("tcp4", addr)
	if err != nil {
		return nil, err
	}
	server.listen = listener
	server.site, err = initSite(s)
	if err != nil {
		return nil, err
	}
	server.db, err = initDb(s)
	if err != nil {
		return nil, err
	}
	return server, nil
}

// 检测配置文件
func (s *Serve) checkConfig() error {
	host, _ := s.main.Conf.GetString("server", "listen")
	port, _ := s.main.Conf.GetString("server", "port")
	if host == "" {
		host = DEF_SERVER_HOST
	}
	if port == "" {
		port = DEF_SERVER_PORT
	}
	serverIV, err := s.main.Conf.GetString("server", "serverIV")
	if err != nil {
		return err
	}
	serverKEY, err := s.main.Conf.GetString("server", "serverKEY")
	if err != nil {
		return err
	}
	serverType, err := s.main.Conf.GetInt("server", "serverType")
	if err != nil {
		return err
	}
	s.host = host
	s.port = port
	s.serverIV = []byte(serverIV)
	s.serverKEY = []byte(serverKEY)
	s.serverType = serverType
	return nil
}

// 使服务器开始服务
func (s *Serve) Accept() {
	s.main.Logger.Println("SFSS server begin serve.")
	for {
		if s.main.Shutdown == true {
			if s.main.ConnNum == 0 {
				s.main.Logger.Println("active connections serve done, now beginning shutdown...")
				break
			} else {
				s.main.Logger.Println("SFSS server has", s.main.ConnNum, "actvie connections, serve continue...")
			}
			time.Sleep(1 * time.Second) // 如果还在处理，等待1000毫秒继续
			continue
		}
		s.main.Logger.Println("SFSS server new accept...")
		conn, err := s.listen.AcceptTCP()
		if err != nil {
			if oerr, ok := err.(*net.OpError); ok && oerr.Err.Error() == "use of closed network connection" {
				/* this hack happens because the error is returned when the
				 * network socket is closing and instead of returning a
				 * io.EOF it returns this error.New(...) struct.
				 */
				continue // 这是listener被关闭，不记录日志
			}
			s.main.Logger.Println(err.Error())
			continue // 输出异常，继续提供服务
		}
		s.main.ConnNum++ // 每启动一个处理，连接数+1
		go s.clientHandle(conn)
	}
	s.main.Logger.Println("SFSS server has been shutdown.")
	s.main.Chs <- 1 // 程序终止，写入Channel数据
}

// 处理客户端请求
func (s *Serve) clientHandle(conn *net.TCPConn) {
	defer func() {
		conn.Close()
		s.main.ConnNum-- // 每结束一个处理，连接数-1
	}()
	var err2 string
	// 接收数据
	data, err := util.TCPConnRead(conn)
	if err != nil {
		err2 = "TCPConnRead Data Error: " + err.Error()
		s.main.Logger.Println(err2)
		s.clientWrite(conn, []byte(err2), 1)
		return
	}
	// 处理接收结构
	receive := new(util.ReceiveData)
	err = json.Unmarshal(data, receive)
	if err != nil {
		err2 = "ReceiveData json decode Error: " + err.Error()
		s.main.Logger.Println(err2)
		s.clientWrite(conn, []byte(err2), 1)
		return
	}
	// 解密数据
	data, err = util.AesDecrypt([]byte(receive.Data), s.serverIV, s.serverKEY)
	if err != nil {
		err2 = "ReceiveData AES decrypt Error: " + err.Error()
		s.main.Logger.Println(err2)
		s.clientWrite(conn, []byte(err2), 1)
		return
	}
	// 判断method进行处理
	order := new(util.OrderData)
	err = json.Unmarshal(data, order)
	if err != nil {
		err2 = "OrderData json decode Error: " + err.Error()
		s.main.Logger.Println(err2)
		s.clientWrite(conn, []byte(err2), 1)
		return
	}
	var result string
	var code int
	switch order.Method {
	case "init_test":
		s.initTest(conn)
		return
	case "site_create":
		result, err = s.site.Create(order.Data)
	case "site_update":
		result, err = s.site.Update(order.Data)
	case "site_pause":
		result, err = s.site.Pause(order.Data)
	case "site_start":
		result, err = s.site.Start(order.Data)
	case "site_delete":
		result, err = s.site.Delete(order.Data)
	case "db_create":
		result, err = s.db.Create(order.Data)
	case "db_update":
		result, err = s.db.Update(order.Data)
	case "db_pause":
		result, err = s.db.Pause(order.Data)
	case "db_start":
		result, err = s.db.Start(order.Data)
	case "db_delete":
		result, err = s.db.Delete(order.Data)
	default:
		result = "method undefined"
		code = 1
	}
	if err != nil {
		s.main.Logger.Println(err.Error())
		result = "handel method " + order.Method + " Error: " + err.Error()
		code = 1
	}
	s.clientWrite(conn, []byte(result), code)
}

// 响应客户端信息
func (s *Serve) clientWrite(conn *net.TCPConn, msg []byte, code int) {
	send := new(util.SendData)
	send.Code = code
	send.Message = string(msg)
	json, _ := json.Marshal(send)
	conn.Write(json)
}

// 停止服务
func (s *Serve) Close() {
	s.listen.Close()
}

// 测试数据方法
func (s *Serve) initTest(conn *net.TCPConn) {
	send := new(util.InitTestData)
	send.Data = make(map[string]interface{})
	send.Data["serverType"] = s.serverType
	json, _ := json.Marshal(send)
	conn.Write(json)
}

// 测试：空方法
func (s *Serve) empty(data map[string]string) (string, error) {
	return "i am empty", nil
}
