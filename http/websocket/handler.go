package websocket

import (
	"github.com/gorilla/websocket"
	"github.com/joeyCheek888/go-std.git/log"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func Handler(w http.ResponseWriter, req *http.Request) {

	// 升级协议
	conn, err := (&websocket.Upgrader{
		HandshakeTimeout: 0,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		WriteBufferPool:  nil,
		Subprotocols:     nil,
		Error:            nil,
		CheckOrigin: func(r *http.Request) bool {
			log.Logger.Info("升级协议",
				zap.String("url", req.URL.String()),
				zap.String("host", req.Host),
				zap.Any("ua", r.Header["User-Agent"]),
				zap.Any("referer", r.Header["Referer"]))
			return true
		},
		EnableCompression: false,
	}).Upgrade(w, req, nil)
	if err != nil {
		http.NotFound(w, req)
		return
	}

	token := req.PathValue("token")
	if token == "" {
		http.NotFound(w, req)
		return
	}
	log.Logger.Info("webSocket 建立连接:", zap.String("addr", conn.RemoteAddr().String()))
	currentTime := time.Now().Unix()
	c := NewClient(conn.RemoteAddr().String(), token, conn, currentTime)
	go c.Read()
	go c.Write()

	// 用户连接事件
	Manager.Register <- c
}

type registerCallbackFun func(client *Client)

var registerCallback = func(client *Client) {
	log.Logger.Info("用户连接", zap.String("addr", client.Addr))
}

func SetRegisterCallback(callback registerCallbackFun) {
	registerCallback = callback
}

type heartbeatCallbackFun func(client *Client)

var heartbeatCallback = func(client *Client) {
	log.Logger.Info(
		"心跳",
		zap.String("addr", client.Addr),
		zap.String("时间", time.Unix(client.HeartbeatTime, 0).Format(time.DateTime)),
	)
}

func SetHeartbeatCallback(callback heartbeatCallbackFun) {
	heartbeatCallback = callback
}

type processMessageCallbackFun func(client *Client, cmd string, data any)

var processMessageCallback = func(client *Client, cmd string, data any) {
	log.Logger.Info(
		"收到消息",
		zap.String("addr", client.Addr),
		zap.Any("clientID", client.ClientID),
		zap.String("cmd", cmd), zap.Any("data", data),
	)
}

func SetProcessMessageCallback(callback processMessageCallbackFun) {
	processMessageCallback = callback
}

type CloseConnCallback func(client *Client)

var closeConnCallback = func(client *Client) {
	log.Logger.Info("关闭连接", zap.String("addr", client.Addr))
}

func SetCloseConnCallback(callback CloseConnCallback) {
	closeConnCallback = callback
}
