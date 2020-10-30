package main

import (
	"MSS_Project/logger"
	"bytes"
	"errors"
	"os/exec"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

const (
	UNRUNNING int32 = 0
	RUNNING   int32 = 1
)

const (
	SUCCESS string = "0"
	FAILED  string = "-1"
	MIDDLE  string = "1"
)

var STATUS int32 = UNRUNNING
var ExprId int32 = -1
var cmdStr string = "../gatling/bin/gatling.sh" // 流量发送的指令

func main() {
	r := gin.Default()
	// 流量发送模块 begin=================
	// 流量发送API

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.POST("/startFlowExpr", startFlowExpr)
	r.POST("/getResult", startFlowExpr)
	// 流量发送模块 end=================
	r.Run("192.168.0.171:58888")
}

func startFlowExpr(c *gin.Context) {
	if atomic.CompareAndSwapInt32(&STATUS, UNRUNNING, RUNNING) == false {
		logger.Error.Println("当前处于实验的RUNNING阶段中：", STATUS)
		c.JSON(200, gin.H{
			"errMsg":    FAILED,
			"ExprIndex": -1,
		})
		return
	}

	// 获得参数
	flowType := c.PostForm("flowType")
	var flowTypeIdx int32
	var err error
	flowTypeIdx, err = CheckPara(flowType)
	if err != nil {
		logger.Error.Println("参数Error：", err)
		c.JSON(200, gin.H{
			"errMsg":    FAILED,
			"ExprIndex": -1,
		})
		return
	}

	// 发送流量
	// 将持续时间参数传给脚本
	// 启动发送流量脚本
	var outInfo bytes.Buffer
	cmd := exec.Command(cmdStr)
	cmd.Stdin = strings.NewReader(flowType + "\n\n")
	cmd.Stdout = &outInfo
	err = cmd.Start()

	ExprId++
	logger.Trace.Println("实验", ExprId, "开始，波形:", ":", flowTypeIdx)

	c.JSON(200, gin.H{
		"errMsg":    SUCCESS,
		"ExprIndex": ExprId,
	})

	go WaitExpr(cmd)
}

func CheckPara(flowType string) (int32, error) {
	var flowTypeIdx int64
	var err error
	flowTypeIdx, err = strconv.ParseInt(flowType, 10, 32)
	if err != nil {
		logger.Error.Println("流量波形参数格式有误" + flowType)
		return -1, errors.New("流量波形参数非整型，格式错误")
	}

	if flowTypeIdx < 0 || flowTypeIdx >= 6 {
		logger.Error.Println("流量波形参数数值有误", flowTypeIdx)
		return -1, errors.New("流量波形参数超出范围，数值错误")
	}
	return (int32)(flowTypeIdx), nil
}

func WaitExpr(cmd *exec.Cmd) {
	err := cmd.Wait() // 阻塞执行
	if err != nil {
		logger.Error.Println("流量发送脚本执行异常 ERROR:", err)
		STATUS = UNRUNNING
		return
	}
	STATUS = UNRUNNING
}

func getResult(c *gin.Context) {
	if STATUS == UNRUNNING && ExprId == -1 {
		c.JSON(200, gin.H{
			"errMsg": FAILED,
			"Msg":    "No history experiment",
		})
		return
	}

	// 获得参数
	var exprStr string
	exprStr = c.PostForm("exprId")

	var exprId int64
	var err error
	exprId, err = strconv.ParseInt(exprStr, 10, 32)
	if err != nil {
		c.JSON(200, gin.H{
			"errMsg": FAILED,
			"Msg":    "Expr Id 格式错误",
		})
		logger.Error.Println("Expr Id 格式错误", exprStr)
		return
	}
	if (int32)(exprId) != ExprId {
		c.JSON(200, gin.H{
			"errMsg": FAILED,
			"Msg":    "Not current experiment",
		})
		logger.Error.Println("不是当前实验:", exprId, "，当前实验为:", ExprId)
		return
	}

	if STATUS == RUNNING {
		c.JSON(200, gin.H{
			"errMsg": MIDDLE,
			"Msg":    "Running",
		})
		logger.Trace.Println("尝试获取实验", exprId, "的状态，当前实状态为: RUNNING")
		return
	}
	c.JSON(200, gin.H{
		"errMsg": SUCCESS,
		"Msg":    "FINISHED",
	})
	logger.Trace.Println("尝试获取实验", exprId, "的状态，当前实状态为: FINISHED")
	return
}
