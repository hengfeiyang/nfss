// Copyright (c) 2013 Beijing CmsTop Technology Co.,Ltd. (http://www.cmstop.com)

// Provides site management
/*
站点管理
开站模板nginx.tpl中一共有四个变量，分别是：[DOMAIN] [ALIAS] [ROOT] [LOG]
*/

package server

import (
	"bufio"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"sfss/util"
	"strings"
)

// 站点操作数据字段：创建
var fieldSiteCreate = [6]string{
	"siteid",      // 站点编号
	"domain",      // 站点主域名
	"alias",       // 站点别名
	"root",        // 站点目录
	"connections", // 站点连接数
	"bandwidth",   // 站点带宽限制
}

// 站点操作数据字段：更新
var fieldSiteUpdate = fieldSiteCreate

// 站点操作数据字段：暂停
var fieldSitePause = [1]string{"domain"}

// 站点操作数据字段：开启
var fieldSiteStart = fieldSitePause

// 站点操作数据字段：删除
var fieldSiteDelete = [2]string{"domain", "root"}

type site struct {
	main         *util.SFSS // 系统接口
	nginxBin     string     // Nginx执行程序
	nginxConfDir string     // Nginx配置文件路径
	siteTpl      string     // 站点配置模板
	siteDir      string     // 站点存储根路径
	logDir       string     // 站点日志存储根路径

}

// 初始化
func initSite(s *util.SFSS) (*site, error) {
	site := new(site)
	site.main = s
	err := site.checkConfig()
	if err != nil {
		return nil, errors.New("checkConfig Error: " + err.Error())
	}
	// 加载站点模板
	dir, err := util.GetDir()
	if err != nil {
		return nil, errors.New("GetDir Error: " + err.Error())
	}
	var tplFile string = dir + "/conf/nginx.tpl"
	tpl, err := ioutil.ReadFile(tplFile)
	if err != nil {
		return nil, errors.New("Read SiteTpl Error: " + err.Error())
	}
	site.siteTpl = string(tpl)
	return site, nil
}

// 检测配置文件
func (s *site) checkConfig() error {
	nginxBin, err := s.main.Conf.GetString("site", "nginxBin")
	if err != nil {
		return err
	}
	nginxConfDir, err := s.main.Conf.GetString("site", "nginxConfDir")
	if err != nil {
		return err
	}
	siteDir, err := s.main.Conf.GetString("site", "siteDir")
	if err != nil {
		return err
	}
	logDir, err := s.main.Conf.GetString("site", "logDir")
	if err != nil {
		return err
	}
	s.nginxBin = nginxBin
	s.nginxConfDir = nginxConfDir
	s.siteDir = siteDir
	s.logDir = logDir
	return nil
}

// 添加站点
func (s *site) Create(data map[string]string) (msg string, err error) {
	var ok bool
	var k, v, config, configFile string

	// 判断站点是否已经存在
	if v, ok = data["domain"]; !ok || v == "" {
		return "", errors.New("domain is empty")
	}
	configFile = s.nginxConfDir + data["domain"] + ".conf"
	ok, err = util.IsExist(configFile)
	// 如果已经存在，直接返回成功
	if ok == true {
		return "site is already exists!", nil
	}

	// 开始处理站点配置文件
	config = s.siteTpl
	for _, k = range fieldSiteCreate {
		if v, ok = data[k]; !ok || v == "" {
			if k == "alias" {
				data[k] = ""
			} else {
				return "", errors.New(k + " is empty")
			}
		}
		if k == "root" {
			data[k] = s.siteDir + data[k]
		}
		config = strings.Replace(config, "["+strings.ToUpper(k)+"]", data[k], 1)
	}
	// 日志单独处理
	log := s.logDir + data["domain"] + "_access.log"
	config = strings.Replace(config, "[LOG]", log, 1)

	// 写入配置文件
	err = ioutil.WriteFile(configFile, []byte(config), 0664)
	if err != nil {
		return "", errors.New("Nginx Config Write Error!" + err.Error())
	}

	// 创建站点目录
	err = os.Mkdir(data["root"], 0755)
	if err != nil {
		return "", errors.New("Site dir create failed!" + err.Error())
	}
	// 设置站点目录权限

	// 重截Nginx使配置变更生效
	var argv []string
	if strings.Index(s.nginxBin, " ") >= 0 {
		argv = strings.Split(s.nginxBin, " ")
		s.nginxBin = argv[0]
		argv = argv[1:]
	} else {
		argv = make([]string, 0)
	}
	cmd := exec.Command(s.nginxBin, argv...)
	_, err = cmd.Output()
	if err != nil {
		return "", errors.New("Nginx reload Error!" + err.Error())
	}

	return "site create ok", nil
}

// 更新站点
func (s *site) Update(data map[string]string) (msg string, err error) {
	var ok bool
	var k, v, config, configFile string

	// 开始处理站点配置文件
	config = s.siteTpl
	for _, k = range fieldSiteUpdate {
		if v, ok = data[k]; !ok || v == "" {
			if k == "alias" {
				data[k] = ""
			} else {
				return "", errors.New(k + " is empty")
			}
		}
		if k == "root" {
			data[k] = s.siteDir + data[k]
		}
		config = strings.Replace(config, "["+strings.ToUpper(k)+"]", data[k], 1)
	}
	// 日志单独处理
	log := s.logDir + data["domain"] + "_access.log"
	config = strings.Replace(config, "[LOG]", log, 1)

	// 写入配置文件
	configFile = s.nginxConfDir + data["domain"] + ".conf"
	err = ioutil.WriteFile(configFile, []byte(config), 0664)
	if err != nil {
		return "", errors.New("Nginx Config Write Error!" + err.Error())
	}

	// 创建站点目录
	ok, _ = util.IsExist(data["root"])
	if ok == false {
		err = os.Mkdir(data["root"], 0755)
		if err != nil {
			return "", errors.New("Site dir create failed!" + err.Error())
		}
	}

	// 设置站点目录权限

	// 重截Nginx使配置变更生效
	var argv []string
	if strings.Index(s.nginxBin, " ") >= 0 {
		argv = strings.Split(s.nginxBin, " ")
		s.nginxBin = argv[0]
		argv = argv[1:]
	} else {
		argv = make([]string, 0)
	}
	cmd := exec.Command(s.nginxBin, argv...)
	_, err = cmd.Output()
	if err != nil {
		return "", errors.New("Nginx reload Error!" + err.Error())
	}

	return "site update ok", nil
}

// 暂停站点
func (s *site) Pause(data map[string]string) (msg string, err error) {
	var ok bool
	var k, v, config, configFile string

	// 开始处理站点配置文件
	// 判断参数
	for _, k = range fieldSitePause {
		if v, ok = data[k]; !ok || v == "" {
			return "", errors.New(k + " is empty")
		}
	}
	configFile = s.nginxConfDir + data["domain"] + ".conf"
	// 读取配置文件
	ok, _ = util.IsExist(configFile)
	if ok == false {
		return "", errors.New("Site " + data["domain"] + " not exist!")
	}
	fileHandle, err := os.OpenFile(configFile, os.O_RDWR, 0664)
	defer fileHandle.Close()
	if err != nil {
		return "", errors.New("Nginx site config file open failed!" + err.Error())
	}
	reader := bufio.NewReader(fileHandle)
	config = ""
	// 注释掉配置文件
	var i int = 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return "", errors.New("Nginx site config read error!" + err.Error())
			}
		}
		if i == 0 && len(line) > 0 && line[0] == '#' {
			return "", errors.New("site already paused!")
		}
		config += "#" + line
		i++
	}
	// 写入配置文件
	_, err = fileHandle.WriteAt([]byte(config), 0)
	if err != nil {
		return "", errors.New("Nginx Config Write Error!" + err.Error())
	}

	// 重截Nginx使配置变更生效
	var argv []string
	if strings.Index(s.nginxBin, " ") >= 0 {
		argv = strings.Split(s.nginxBin, " ")
		s.nginxBin = argv[0]
		argv = argv[1:]
	} else {
		argv = make([]string, 0)
	}
	cmd := exec.Command(s.nginxBin, argv...)
	_, err = cmd.Output()
	if err != nil {
		return "", errors.New("Nginx reload Error!" + err.Error())
	}

	return "site pause ok", nil
}

// 开启站点
func (s *site) Start(data map[string]string) (msg string, err error) {
	var ok bool
	var k, v, config, configFile string

	// 开始处理站点配置文件
	// 判断参数
	for _, k = range fieldSiteStart {
		if v, ok = data[k]; !ok || v == "" {
			return "", errors.New(k + " is empty")
		}
	}
	configFile = s.nginxConfDir + data["domain"] + ".conf"
	// 读取配置文件
	ok, _ = util.IsExist(configFile)
	if ok == false {
		return "", errors.New("Site " + data["domain"] + " not exist!")
	}
	fileHandle, err := os.OpenFile(configFile, os.O_RDWR, 0664)
	defer fileHandle.Close()
	if err != nil {
		return "", errors.New("Nginx site config file open failed!" + err.Error())
	}
	reader := bufio.NewReader(fileHandle)
	config = ""
	// 清理注释配置文件
	var i int = 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return "", errors.New("Nginx site config read error!" + err.Error())
			}
		}
		if i == 0 && len(line) > 0 && line[0] != '#' {
			return "", errors.New("site already started!")
		}
		config += string(line[1:])
		i++
	}
	// 写入配置文件
	fileHandle.Truncate(0)
	_, err = fileHandle.WriteAt([]byte(config), 0)
	if err != nil {
		return "", errors.New("Nginx Config Write Error!" + err.Error())
	}

	// 重截Nginx使配置变更生效
	var argv []string
	if strings.Index(s.nginxBin, " ") >= 0 {
		argv = strings.Split(s.nginxBin, " ")
		s.nginxBin = argv[0]
		argv = argv[1:]
	} else {
		argv = make([]string, 0)
	}
	cmd := exec.Command(s.nginxBin, argv...)
	_, err = cmd.Output()
	if err != nil {
		return "", errors.New("Nginx reload Error!" + err.Error())
	}

	return "site start ok", nil
}

// 删除站点
func (s *site) Delete(data map[string]string) (msg string, err error) {
	var ok bool
	var k, v, configFile string

	// 开始处理站点配置文件
	// 判断参数
	for _, k = range fieldSiteDelete {
		if v, ok = data[k]; !ok || v == "" {
			return "", errors.New(k + " is empty")
		}
		if k == "root" {
			data[k] = s.siteDir + data[k]
		}
	}
	configFile = s.nginxConfDir + data["domain"] + ".conf"
	// 删除配置文件
	err = os.Remove(configFile)
	if err != nil {
		if _, ok := err.(*os.PathError); !ok {
			return "", errors.New("Nginx site config delete Error!" + err.Error())
		}
	}

	// 重截Nginx使配置变更生效
	var argv []string
	if strings.Index(s.nginxBin, " ") >= 0 {
		argv = strings.Split(s.nginxBin, " ")
		s.nginxBin = argv[0]
		argv = argv[1:]
	} else {
		argv = make([]string, 0)
	}
	cmd := exec.Command(s.nginxBin, argv...)
	_, err = cmd.Output()
	if err != nil {
		return "", errors.New("Nginx reload Error!" + err.Error())
	}
	// 删除站点目录
	os.Rename(data["root"], data["root"]+".bak")

	// 删除站点日志
	log := s.logDir + data["domain"] + "_access.log"
	err = os.Remove(log)
	if err != nil {
		if _, ok := err.(*os.PathError); !ok {
			return "", errors.New("Nginx site logfile delete Error!" + err.Error())
		}
	}

	return "site delete ok", nil
}
