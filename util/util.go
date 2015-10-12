// Copyright (c) 2013 Beijing CmsTop Technology Co.,Ltd. (http://www.cmstop.com)

/*
Package util provides support for format and read SFSS protocol.
*/
package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

const (
	MAX_PACKET_SIZE = 1048576                                                          // 数据包最大为1M
	RAND_CHARS      = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ" // 随机字符串的母本
	RAND_MAX_LENGTH = 64                                                               // 随机字符串最大长度
)

// 协议封装读取
func TCPConnRead(conn *net.TCPConn) ([]byte, error) {
	conn.SetReadDeadline(time.Now().Add(time.Second * 30))
	result := bytes.NewBuffer(nil)
	data := make([]byte, 4)
	num, err := conn.Read(data)
	if err != nil || num != 4 {
		if err == nil {
			err = errors.New("length read error")
		}
		return nil, err
	}
	result.Write(data[0:num])
	var length int32
	err = binary.Read(result, binary.LittleEndian, &length)
	if err != nil {
		return nil, err
	}
	if length > MAX_PACKET_SIZE {
		return nil, errors.New("too large packet! packet size should less than 1M")
	}
	data = make([]byte, length)
	result = bytes.NewBuffer(nil)
	num, err = io.ReadFull(conn, data)
	if err != nil {
		return nil, err
	}
	result.Write(data[0:num])
	return result.Bytes(), nil
}

// aes加密
func AesEncrypt(data, iv, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	data = PKCS5Padding(data, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, iv)
	cryptData := make([]byte, len(data))
	blockMode.CryptBlocks(cryptData, data)
	baseData := make([]byte, base64.StdEncoding.EncodedLen(len(cryptData)))
	base64.StdEncoding.Encode(baseData, cryptData)
	return baseData, nil
}

// aes解密
func AesDecrypt(data, iv, key []byte) ([]byte, error) {
	baseData := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
	length, err := base64.StdEncoding.Decode(baseData, data)
	if err != nil {
		return nil, err
	}
	data = baseData[:length]
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, iv)
	blockSize := block.BlockSize()
	origData := make([]byte, len(data))
	blockMode.CryptBlocks(origData, data)
	origData, err = PKCS5UnPadding(origData, blockSize)
	if err != nil {
		return nil, err
	}
	return origData, nil
}

// aes加密补码
func PKCS5Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

// aes解密去码
func PKCS5UnPadding(data []byte, blockSize int) ([]byte, error) {
	length := len(data)
	unpadding := int(data[length-1])
	if unpadding >= blockSize {
		return nil, errors.New("AES PCKS5UnPadding penic, unpadding Illegal")
	}
	return data[:(length - unpadding)], nil
}

// 获取程序运行的目录
func GetDir() (string, error) {
	path, err := filepath.Abs(os.Args[0])
	if err != nil {
		return "", err
	}
	return filepath.Dir(path), nil
}

// 判断一个文件或目录是否存在
func IsExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	// Check if error is "no such file or directory"
	if _, ok := err.(*os.PathError); ok {
		return false, nil
	}
	return false, err
}

// 判断一个文件或目录是否有写入权限
func IsWritable(path string) (bool, error) {
	err := syscall.Access(path, syscall.O_RDWR)
	if err == nil {
		return true, nil
	}
	// Check if error is "no such file or directory"
	if _, ok := err.(*os.PathError); ok {
		return false, nil
	}
	return false, err
}

// 生成一个随机字符串
func RandString(n int) string {
	if n > RAND_MAX_LENGTH {
		n = RAND_MAX_LENGTH
	}
	s := make([]byte, 0, n+1)
	sn := len(RAND_CHARS)
	rand.Seed(int64(time.Now().Nanosecond()))
	for i := 0; i < n; i++ {
		s = append(s, RAND_CHARS[rand.Intn(sn)])
	}
	return string(s)
}

// 格式化size单位 输出友好格式(B,KB,MB,GB,TB)
func FormatSize(s int64) string {
	if s >= 1<<40 {
		f := float64(s) / (1 << 40)
		return fmt.Sprintf("%.2f TB", f)
	}
	if s >= 1<<30 {
		f := float64(s) / (1 << 30)
		return fmt.Sprintf("%.2f GB", f)
	}
	if s >= 1<<20 {
		f := float64(s) / (1 << 20)
		return fmt.Sprintf("%.2f MB", f)
	}
	if s >= 1<<10 {
		f := float64(s) / (1 << 10)
		return fmt.Sprintf("%.2f KB", f)
	}
	return fmt.Sprintf("%d byte", s)
}

// 读取一个文件夹返回文件列表
func ReadDir(dirname string) ([]os.FileInfo, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	return list, nil
}

// 获取一个文件或目录的大小
func GetPathSize(path string) (int64, error) {
	var size int64 = 0
	s, err := os.Stat(path)
	if err != nil {
		return size, err
	}
	if s.IsDir() == true {
		fl, err := ReadDir(path)
		if err != nil {
			return size, err
		}
		for _, v := range fl {
			vs, err := GetPathSize(path + "/" + v.Name())
			if err != nil {
				return size, err
			}
			size += vs
		}
	} else {
		size += s.Size()
	}
	return size, nil
}
