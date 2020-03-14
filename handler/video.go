package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/lixiangyuan98/ogateway/conf"
	"github.com/lixiangyuan98/ogateway/control/video"
	"github.com/lixiangyuan98/ogateway/model"
	"github.com/lixiangyuan98/ogateway/util"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func Send(ctx *gin.Context) {
	req := &model.VideoSendRequest{}
	if err := ctx.Bind(req); err != nil {
		log.Printf("bind error: %v", err)
		ctx.JSON(400, gin.H{
			"message": "invalid parameter",
		})
		return
	}
	port := strconv.FormatInt(int64(req.Port), 10)
	err := video.Send(req.Method, req.Host, req.Src, req.Dest, port)
	if err != nil {
		log.Printf("send error: %v", err)
		ctx.JSON(400, gin.H{
			"message": "send error",
		})
		return
	}
	ctx.JSON(200, gin.H{
		"message": "ok",
	})
}

func ConnectVideoServer(ctx *gin.Context) {
	ws, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Printf("upgrade to websocket error: %v\n", err)
		return
	}
	defer func() {
		log.Printf("ws closed by remote\n")
		ws.Close()
	}()
	mu := &sync.Mutex{}

	closed := make(chan struct{})
	timer := time.NewTimer(conf.GlobalConf.VideoConf.GetInfoInterval)
	user, _ := ctx.Get("user")
	hosts := make([]string, 0)
	for _, host := range conf.VideoConf {
		if conf.UserInGroup(user.(*conf.User), host.Group) {
			hosts = append(hosts, host.Host)
		}
	}

	go func() {
		for {
			select {
			case <-closed:
				log.Printf("close ws://%v\n", ws.RemoteAddr().String())
				return
			case <-timer.C:
				servers := video.GetInfo(hosts)
				mu.Lock()
				util.WriteJson(ws, 0, "info", servers)
				mu.Unlock()
				timer.Reset(conf.GlobalConf.VideoConf.GetInfoInterval)
			}
		}
	}()
	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			log.Printf("read message from ws://%v error: %v\n", ws.RemoteAddr().String(), err)
			closed <- struct{}{}
			return
		}
		v := model.VideoSendRequest{}
		if err := json.Unmarshal(message, &v); err != nil {
			log.Printf("resolve VideoSendRequest error: %v\n", err)
			mu.Lock()
			util.WriteNil(ws, -201, "invalid request")
			mu.Unlock()
			continue
		}
		port := strconv.FormatInt(int64(v.Port), 10)
		if err := video.Send(v.Method, v.Host, v.Src, v.Dest, port); err != nil {
			mu.Lock()
			util.WriteNil(ws, -300, err.Error())
			mu.Unlock()
			continue
		}
		mu.Lock()
		util.WriteNil(ws, 0, "ok")
		mu.Unlock()
	}
}
