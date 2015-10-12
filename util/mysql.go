// Copyright (c) 2013 Beijing CmsTop Technology Co.,Ltd. (http://www.cmstop.com)

// Provides some mysql methods
package util

import (
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"strings"
)

const (
	DENI_DB_NAMES = "mysql,test,admin" // 禁止使用的数据库
)

// 数据库连接结构
type DbMySQL struct {
	Psn  string
	Conn *sql.DB
}

// 创建数据库实例
func NewDb(psn string) (*DbMySQL, error) {
	if len(psn) < 10 {
		return nil, errors.New("PSN is empty!")
	}
	ndb := new(DbMySQL)
	ndb.Psn = psn
	conn, err := ndb.connect()
	if err != nil {
		return nil, err
	}
	ndb.Conn = conn
	return ndb, nil
}

func (s *DbMySQL) connect() (*sql.DB, error) {
	conn, err := sql.Open("mysql", s.Psn)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (s *DbMySQL) ping() error {
	err := s.Conn.Ping()
	if err != nil {
		// 重新连接
		conn, err := s.connect()
		if err != nil {
			return err
		}
		s.Conn = conn
	}
	return nil
}

// 获取数据库版本
func (s *DbMySQL) Version() (string, error) {
	err := s.ping()
	if err != nil {
		return "", err
	}
	row := s.Conn.QueryRow("SELECT VERSION()")
	var v string
	err = row.Scan(&v)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("query result is empty")
		}
		return "", err
	}
	return v, nil
}

// 执行一条SQL语句
// 如果出错返回err,正常返回nil
func (s *DbMySQL) Exec(sql string) error {
	var err error
	err = s.ping()
	if err != nil {
		return err
	}
	_, err = s.Conn.Exec(sql)
	if err != nil {
		return err
	}
	return nil
}

// 刷新数据库权限
func (s *DbMySQL) Flush() error {
	var err error
	err = s.ping()
	if err != nil {
		return err
	}
	_, err = s.Conn.Exec("FLUSH PRIVILEGES")
	if err != nil {
		return err
	}
	return nil
}

// 检测数据库名称
// 检测通过（可用）为true，检测失败为false
func (s *DbMySQL) checkDbName(name string) bool {
	name = strings.ToLower(name)
	dn := strings.Split(DENI_DB_NAMES, ",")
	for _, v := range dn {
		if name == v {
			return false
		}
	}
	return true
}

// 创建数据库
func (s *DbMySQL) CreateDb(name string) error {
	var err error
	err = s.ping()
	if err != nil {
		return err
	}
	if ok := s.checkDbName(name); !ok {
		return errors.New("name is deny!")
	}
	_, err = s.Conn.Exec("CREATE DATABASE IF NOT EXISTS " + name + " default charset utf8 COLLATE utf8_general_ci")
	if err != nil {
		return err
	}
	return nil
}

// 创建用户并设置权限
func (s *DbMySQL) CreateUser(name, user, host, pass string) error {
	var err error
	err = s.ping()
	if err != nil {
		return err
	}
	if ok := s.checkDbName(name); !ok {
		return errors.New("name is deny!")
	}
	_, err = s.Conn.Exec("GRANT ALL ON " + name + ".* TO '" + user + "'@'" + host + "' IDENTIFIED BY '" + pass + "'")
	if err != nil {
		return err
	}
	return nil
}

// 修改用户密码
func (s *DbMySQL) Password(user, pass string) error {
	var err error
	err = s.ping()
	if err != nil {
		return err
	}
	_, err = s.Conn.Exec("UPDATE mysql.user SET `password`=PASSWORD('" + pass + "') WHERE user='" + user + "'")
	if err != nil {
		return err
	}
	return nil
}

// 删除数据库
func (s *DbMySQL) DeleteDb(name string) error {
	var err error
	err = s.ping()
	if err != nil {
		return err
	}
	if ok := s.checkDbName(name); !ok {
		return errors.New("name is deny!")
	}
	_, err = s.Conn.Exec("DROP DATABASE IF EXISTS " + name)
	if err != nil {
		return err
	}
	return nil
}

// 删除用户
func (s *DbMySQL) DeleteUser(name string) error {
	var err error
	err = s.ping()
	if err != nil {
		return err
	}
	if ok := s.checkDbName(name); !ok {
		return errors.New("name is deny!")
	}
	_, err = s.Conn.Exec("DELETE FROM mysql.user WHERE user='" + name + "'")
	if err != nil {
		return err
	}
	return nil
}

// 获取数据库大小
func (s *DbMySQL) GetDbSize(name string) (int64, error) {
	var size int64
	var err error
	err = s.ping()
	if err != nil {
		return size, err
	}
	if ok := s.checkDbName(name); !ok {
		return size, errors.New("name is deny!")
	}
	row := s.Conn.QueryRow("SELECT SUM(DATA_LENGTH) + SUM(INDEX_LENGTH) FROM information_schema.TABLES WHERE table_schema='" + name + "'")
	err = row.Scan(&size)
	if err != nil {
		if err == sql.ErrNoRows {
			return size, errors.New("query result is empty")
		}
		return size, err
	}
	return size, nil
}
