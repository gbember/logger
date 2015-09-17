//日志功能
//日志文件格式 info_log_yyyy_mm_dd.log
//每日凌晨0分1秒创建当天日志文件
//设置log标准库写向当前日志文件
package logger

import (
	"errors"
	"fmt"
	glog "log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

var (
	log *_logger
	mt  sync.Mutex

	DebugLogFun    = func(s string) { Debug(s) }
	InfoLogFun     = func(s string) { Info(s) }
	ErrorLogFun    = func(s string) { Error(s) }
	CriticalLogFun = func(s string) { Critical(s) }
)

const (
	CRITICAL int = iota + 1
	ERROR
	INFO
	DEBUG
)

type _logger struct {
	fd       *os.File
	logChan  chan *string
	logLevel int
	logDir   string
	timer    *time.Timer
}

//启动日志
//dir 日志文件存放目录
//logLevel 日志等级
func StartLog(dir string, logLevel int) error {
	mt.Lock()
	defer mt.Unlock()
	if log != nil {
		return errors.New("文件日志已经启动")
	}
	tlog := &_logger{logLevel: logLevel, logDir: dir}
	err := tlog.run()
	if err != nil {
		return errors.New(fmt.Sprintf("启动日志错误:%s", err.Error()))
	}
	log = tlog
	return nil
}

//调试日志
func Debug(format string, args ...interface{}) {
	if log != nil {
		log.debug(format, args...)
	}
}

//信息日志
func Info(format string, args ...interface{}) {
	if log != nil {
		log.info(format, args...)
	}
}

//错误日志
func Error(format string, args ...interface{}) {
	if log != nil {
		log.error(format, args...)
	}
}

//系统日志
func Critical(format string, args ...interface{}) {
	if log != nil {
		log.critical(format, args...)
	}
}

func (l *_logger) debug(format string, args ...interface{}) {
	if l.logLevel >= DEBUG {
		str := fmt.Sprintf(format, args...)
		pc, file, lineno, ok := runtime.Caller(2)
		src := ""
		if ok {
			src = fmt.Sprintf("%s[DEBUG](%s=[%s]:%d) %s\n", time.Now().Format("======2006/01/02 15:04:05====="),
				runtime.FuncForPC(pc).Name(), filepath.Base(file), lineno, str)
		} else {
			src = fmt.Sprintf("%s[DEBUG] %s\n", time.Now().Format("======2006/01/02 15:04:05====="), str)
		}
		l.logChan <- &src
	}
}
func (l *_logger) info(format string, args ...interface{}) {
	if l.logLevel >= INFO {
		str := fmt.Sprintf(format, args...)
		pc, file, lineno, ok := runtime.Caller(2)
		src := ""
		if ok {
			src = fmt.Sprintf("%s[INFO](%s=[%s]:%d) %s\n", time.Now().Format("======2006/01/02 15:04:05====="),
				runtime.FuncForPC(pc).Name(), filepath.Base(file), lineno, str)
		} else {
			src = fmt.Sprintf("%s[INFO] %s\n", time.Now().Format("======2006/01/02 15:04:05====="), str)
		}
		l.logChan <- &src
	}
}
func (l *_logger) error(format string, args ...interface{}) {
	if l.logLevel >= ERROR {
		str := fmt.Sprintf(format, args...)
		pc, file, lineno, ok := runtime.Caller(2)
		src := ""
		if ok {
			src = fmt.Sprintf("%s[Error](%s=[%s]:%d) %s\n", time.Now().Format("======2006/01/02 15:04:05====="),
				runtime.FuncForPC(pc).Name(), filepath.Base(file), lineno, str)
		} else {
			src = fmt.Sprintf("%s[Error] %s\n", time.Now().Format("======2006/01/02 15:04:05====="), str)
		}
		l.logChan <- &src
	}
}
func (l *_logger) critical(format string, args ...interface{}) {
	if l.logLevel >= CRITICAL {
		str := fmt.Sprintf(format, args...)
		pc, file, lineno, ok := runtime.Caller(2)
		src := ""
		if ok {
			src = fmt.Sprintf("%s[CRITICAL](%s=[%s]:%d) %s\n", time.Now().Format("======2006/01/02 15:04:05====="),
				runtime.FuncForPC(pc).Name(), filepath.Base(file), lineno, str)
		} else {
			src = fmt.Sprintf("%s[CRITICAL] %s\n", time.Now().Format("======2006/01/02 15:04:05====="), str)
		}
		l.logChan <- &src
	}
}

//启动日志
func (l *_logger) run() error {
	err := os.MkdirAll(l.logDir, 0660)
	if err != nil {
		return err
	}
	var logFilename string = getLogInfoFileName()
	logFile := filepath.Join(l.logDir, logFilename)
	fd, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	logChan := make(chan *string, 100000)
	l.fd = fd
	l.logChan = logChan
	glog.SetOutput(fd)
	glog.SetFlags(glog.Llongfile | glog.LstdFlags)
	glog.SetPrefix("[log] ")
	go l.loop()
	return nil
}

func (l *_logger) loop() {
	l.changeFDTimer()
	defer func() {
		if l.timer != nil {
			l.timer.Stop()
		}
	}()
	for {
		select {
		case s, ok := <-l.logChan:
			if ok {
				l.fd.WriteString(*s)
			} else {
				return
			}
		case <-l.timer.C:
			l.changeFDTimer()
			var logFilename string = getLogInfoFileName()
			logFile := filepath.Join(l.logDir, logFilename)
			fd, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				l.error("变换日志文件错误:%s", err.Error())
			} else {
				glog.SetOutput(fd)
				l.fd.Close()
				l.fd = fd
			}

		}
	}
}

func (l *_logger) changeFDTimer() {
	if l.timer != nil {
		l.timer.Stop()
	}
	tNow := time.Now()
	tt := time.Date(tNow.Year(), tNow.Month(), tNow.Day(), 0, 0, 1, 0, time.UTC)
	if tNow.Unix() >= tt.Unix() {
		tt = tt.AddDate(0, 0, 1)
	}
	l.timer = time.NewTimer(tt.Sub(tNow))
}

//根据日期得到log文件
func getLogInfoFileName() string {
	year, mouth, day := time.Now().Date()
	return fmt.Sprintf("info_log_%d_%d_%d.log", year, int(mouth), day)
}
