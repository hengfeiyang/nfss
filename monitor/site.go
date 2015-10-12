package monitor

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"sfss/util"
	"strconv"
)

func getUsage() {
	var data map[string]int64
	data = make(map[string]int64)
	file := "../log/var/httpd-access.log"
	f, err := os.Open(file)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	r := bufio.NewReader(f)
	var i = 0
	for {
		i++
		line, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err.Error())
			os.Exit(1)
		}
		re, _ := regexp.Compile("([0-9a-z.-]+) .*\".*\" [0-9]+ ([0-9]+)$")
		s := re.FindSubmatch(line)
		if len(s) != 3 {
			continue
		}
		key := string(s[1])
		value, err := strconv.ParseInt(string(s[2]), 10, 64)
		if err != nil {
			continue
		}
		if _, ok := data[key]; ok {
			data[key] += value
		} else {
			data[key] = value
		}
	}
	fmt.Println("total: ", i)
	for k, v := range data {
		fmt.Println(k, util.FormatSize(v))
	}
}
