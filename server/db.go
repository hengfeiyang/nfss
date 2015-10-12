// Copyright (c) 2013 Beijing CmsTop Technology Co.,Ltd. (http://www.cmstop.com)

// Provides db management

package server

import (
	"errors"
	"sfss/util"
)

// 数据库操作数据字段：创建
var fieldDbCreate = [4]string{
	"name",     // 数据库名称
	"user",     // 数据库帐号
	"host",     // 数据库帐号主机
	"password", // 数据库帐号密码
}

// 数据库操作数据字段：更新
var fieldDbUpdate = fieldDbCreate

// 数据库操作数据字段：暂停
var fieldDbPause = [1]string{"user"}

// 数据库操作数据字段：开启
var fieldDbStart = [2]string{"user", "password"}

// 数据库操作数据字段：删除
var fieldDbDelete = [2]string{"name", "user"}

type db struct {
	main      *util.SFSS    // 系统接口
	mysqlHost string        // MySQL服务主机
	mysqlPort string        // MySQL服务端口
	mysqlUser string        // MySQL管理帐号
	mysqlPass string        // MySQL管理密码
	backupDir string        // 数据库删除时备份目录
	conn      *util.DbMySQL // 数据库操作接口
}

// 初始化
func initDb(s *util.SFSS) (*db, error) {
	db := new(db)
	db.main = s
	err := db.checkConfig()
	if err != nil {
		return nil, errors.New("checkConfig Error: " + err.Error())
	}
	// 初始化一个MySQL连接，使用keepalive
	conn, err := util.NewDb(db.mysqlUser + ":" + db.mysqlPass + "@tcp(" + db.mysqlHost + ":" + db.mysqlPort + ")/?charset=utf8")
	if err != nil {
		return nil, errors.New("mysql Check Error: " + err.Error())
	}
	db.conn = conn
	return db, nil
}

// 检测配置文件
func (s *db) checkConfig() error {
	host, err := s.main.Conf.GetString("db", "mysqlHost")
	if err != nil {
		return err
	}
	port, err := s.main.Conf.GetString("db", "mysqlPort")
	if err != nil {
		return err
	}
	user, err := s.main.Conf.GetString("db", "mysqlUser")
	if err != nil {
		return err
	}
	pass, err := s.main.Conf.GetString("db", "mysqlPass")
	if err != nil {
		return err
	}
	backupDir, err := s.main.Conf.GetString("db", "backupDir")
	if err != nil {
		return err
	}
	var ok bool
	// 检测备份路径是否存在
	ok, _ = util.IsExist(backupDir)
	if ok == false {
		return errors.New("db backupDir is not exist!")
	}
	// 检测备份路径是否可写
	ok, _ = util.IsWritable(backupDir)
	if ok == false {
		return errors.New("db backupDir is not writable!")
	}
	s.mysqlHost = host
	s.mysqlPort = port
	s.mysqlUser = user
	s.mysqlPass = pass
	s.backupDir = backupDir
	return nil
}

// 添加数据库
func (s *db) Create(data map[string]string) (msg string, err error) {
	var ok bool
	var k, v string
	// 开始处理配置文件
	for _, k = range fieldDbCreate {
		if v, ok = data[k]; !ok || v == "" {
			return "", errors.New(k + " is empty")
		}
	}
	// 创建数据库
	err = s.conn.CreateDb(data["name"])
	if err != nil {
		return "", errors.New("create database error:" + err.Error())
	}

	// 创建用户
	err = s.conn.CreateUser(data["name"], data["user"], data["host"], data["password"])
	if err != nil {
		return "", errors.New("create db user error:" + err.Error())
	}

	// 刷新权限
	err = s.conn.Flush()
	if err != nil {
		return "", errors.New("db flush error:" + err.Error())
	}

	return "db create ok", nil
}

// 更新数据库
func (s *db) Update(data map[string]string) (msg string, err error) {
	var ok bool
	var k, v string
	// 开始处理配置文件
	for _, k = range fieldDbUpdate {
		if v, ok = data[k]; !ok || v == "" {
			return "", errors.New(k + " is empty")
		}
	}
	// 创建数据库
	err = s.conn.CreateDb(data["name"])
	if err != nil {
		return "", errors.New("create database error:" + err.Error())
	}

	// 创建用户
	err = s.conn.CreateUser(data["name"], data["user"], data["host"], data["password"])
	if err != nil {
		return "", errors.New("create db user error:" + err.Error())
	}

	// 刷新权限
	err = s.conn.Flush()
	if err != nil {
		return "", errors.New("db flush error:" + err.Error())
	}

	return "db update ok", nil
}

// 暂停数据库
func (s *db) Pause(data map[string]string) (msg string, err error) {
	var ok bool
	var k, v string
	// 开始处理配置文件
	for _, k = range fieldDbPause {
		if v, ok = data[k]; !ok || v == "" {
			return "", errors.New(k + " is empty")
		}
	}
	// 修改密码
	pass := util.RandString(15)
	err = s.conn.Password(data["user"], pass)
	if err != nil {
		return "", errors.New("stop user error:" + err.Error())
	}

	// 刷新权限
	err = s.conn.Flush()
	if err != nil {
		return "", errors.New("db flush error:" + err.Error())
	}

	return "db pause ok", nil
}

// 开启数据库
func (s *db) Start(data map[string]string) (msg string, err error) {
	var ok bool
	var k, v string
	// 开始处理配置文件
	for _, k = range fieldDbPause {
		if v, ok = data[k]; !ok || v == "" {
			return "", errors.New(k + " is empty")
		}
	}
	// 修改密码
	err = s.conn.Password(data["user"], data["password"])
	if err != nil {
		return "", errors.New("start user error:" + err.Error())
	}

	// 刷新权限
	err = s.conn.Flush()
	if err != nil {
		return "", errors.New("db flush error:" + err.Error())
	}

	return "db start ok", nil
}

// 删除数据库
func (s *db) Delete(data map[string]string) (msg string, err error) {
	var ok bool
	var k, v string
	// 开始处理配置文件
	for _, k = range fieldDbPause {
		if v, ok = data[k]; !ok || v == "" {
			return "", errors.New(k + " is empty")
		}
	}
	// 删除用户
	err = s.conn.DeleteUser(data["user"])
	if err != nil {
		return "", errors.New("delete user error:" + err.Error())
	}

	// 删除数据库
	err = s.conn.DeleteDb(data["name"])
	if err != nil {
		return "", errors.New("delete db error:" + err.Error())
	}

	// 刷新权限
	err = s.conn.Flush()
	if err != nil {
		return "", errors.New("db flush error:" + err.Error())
	}

	return "db delete ok", nil
}
