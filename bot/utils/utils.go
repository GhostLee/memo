package utils

import (
	"log"
	"regexp"
	"runtime"
	"strings"
)

var (
	cqReg = regexp.MustCompile(`\[CQ:.*\]\s`)
	cmdReg = regexp.MustCompile(`^([#][a-z]+)`)
)

func SplitCmd(v string) []string  {
	v = cqReg.ReplaceAllString(v, " ")
	cmds := strings.Split(v, " ")
	if len(cmds)==0{
		return cmds
	}
	if !cmdReg.MatchString(cmds[0]){
		return []string{}
	}
	return cmds
}

func PrintMemStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Printf("Alloc = %vKB TotalAlloc = %vKB Sys = %vKB NumGC = %vKB \n", m.Alloc/1024, m.TotalAlloc/1024, m.Sys/1024, m.NumGC)
}