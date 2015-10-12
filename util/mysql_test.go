package util

import (
	"fmt"
	"testing"
)

func TestVersion(t *testing.T) {
	db, err := NewDb("root:123456@tcp(127.0.0.1:3306)/?charset=utf8")
	if err != nil {
		t.Error(err)
	}
	ver, err := db.Version()
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("MySQL Version is %s\n", ver)
}

func TestExec1(t *testing.T) {
	db, err := NewDb("root:123456@tcp(127.0.0.1:3306)/?charset=utf8")
	if err != nil {
		t.Error(err)
	}
	err = db.Exec("select version()")
	if err != nil {
		t.Error(err)
	}
}

func TestGetDbSize1(t *testing.T) {
	db, err := NewDb("root:123456@tcp(127.0.0.1:3306)/?charset=utf8")
	if err != nil {
		t.Error(err)
	}
	size, err := db.GetDbSize("fy_dongman")
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Db Size %s \n", FormatSize(size))
}
