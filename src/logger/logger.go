package logger

import (
	"io"
	"log"
	"os"
)

// 日志文件相关配置 begin ================
var LogFilePath string = "../log/server.log"

var (
	Trace   *log.Logger // 记录所有日志
	Info    *log.Logger // 重要的信息, 暂不设置
	Warning *log.Logger // 需要注意的信息，暂不设置
	Error   *log.Logger // 非常严重的问题
)

// 日志信息初始化
func LoggerInit() {
	file, err := os.OpenFile(LogFilePath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open error log file:", err)
	}

	Trace = log.New(file,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(file,
		"Info: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(file,
		"Warning: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(io.MultiWriter(file, os.Stderr),
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}
