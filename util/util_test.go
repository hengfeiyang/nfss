// Copyright (c) 2013 Beijing CmsTop Technology Co.,Ltd. (http://www.cmstop.com)

package util

import (
	"fmt"
	"testing"
)

var iv = []byte("1234567890123456")
var key = []byte("1234567890123456")

func TestEncrypt1(t *testing.T) {
	plainText := []byte("i am coldstar")
	cryptText, err := AesEncrypt(plainText, iv, key)
	if err != nil {
		t.Error("encrypt failed: ", err.Error())
	}
	decryptText, err := AesDecrypt(cryptText, iv, key)
	if err != nil {
		t.Error("decrypt failed: ", err.Error())
	}
	if string(decryptText) != string(plainText) {
		t.Error("data check error!")
	}
}

func TestRandString1(t *testing.T) {
	s := RandString(35)
	fmt.Printf("RandString Length 35:\n%s\n", s)
}

func TestGetDirSize1(t *testing.T) {
	d := "/Users/yanghengfei/Desktop"
	s, err := GetPathSize(d)
	if err != nil {
		t.Error("GetPathSize failed: ", err.Error())
	}
	fmt.Printf("GetPathSize %s is %s\n", d, FormatSize(s))
}
