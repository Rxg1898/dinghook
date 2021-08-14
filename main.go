package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	logger "github.com/sirupsen/logrus"
)

type dingHook struct {
	apiUrl     string         // 机器人地址
	levels     []logger.Level // 必带字段
	atMobiles  []string       // at谁
	appName    string         // 模块前缀
	jsonBodies chan []byte    // 异步发送内容
	closeChan  chan bool      // 主进程关闭通道
}

// 代表在哪几个级别下应用这个hook
func (dh *dingHook) Levels() []logger.Level {
	return dh.levels
}

// 执行哪个函数
func (dh *dingHook) Fire(e *logger.Entry) error {
	msg, _ := e.String()
	dh.DirectSend(msg)
	return nil
}

// 同步发送
func (dh *dingHook) DirectSend(msg string) {
	dm := dingMsg{
		MsgType: "text",
	}
	dm.Text.Content = fmt.Sprintf("[日志告警log]\n[app=%s]\n"+
		"[日志详情：%s]", dh.appName, msg)

	dm.At.AtMobiles = dh.atMobiles
	bs, err := json.Marshal(dm)
	if err != nil {
		logger.Errorf("[消息json.marshal失败: %v][msg:%v]", err, msg)
		return
	}
	res, err := http.Post(dh.apiUrl, "application/json", bytes.NewBuffer(bs))
	if err != nil {
		logger.Errorf("[消息发送失败: %v][msg:%v]", err, msg)
		return
	}
	if res != nil && res.StatusCode != 200 {
		logger.Errorf("[顶顶返回错误][StatusCode:%v][msg:%v]", res.StatusCode, msg)
		return
	}

}

// 定义发钉钉信息的字段
type dingMsg struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
	At struct {
		AtMobiles []string `json:"atMobiles"`
	} `json:"at"`
}

func main() {
	dh := &dingHook{
		apiUrl:     "https://oapi.dingtalk.com/robot/send?access_token=xxx",
		levels:     []logger.Level{logger.WarnLevel, logger.InfoLevel},
		atMobiles:  []string{""},
		appName:    "live",
		jsonBodies: make(chan []byte),
		closeChan:  make(chan bool),
	}
	// dh.DirectSend("测试一条看看")
	level := logger.InfoLevel
	logger.SetLevel(level)
	// 设置filename
	logger.SetReportCaller(true)
	logger.SetFormatter(&logger.JSONFormatter{
		TimestampFormat: "2006-01-02 15-04-05",
	})
	// 添加hook
	logger.AddHook(dh)
	logger.Info("这是日志hook的logrus")
}
