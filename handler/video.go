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
	timer := time.NewTimer(0)
	user, _ := ctx.Get("user")
	hosts := make([]string, 0)
	for _, host := range conf.VideoConf {
		if conf.UserInGroup(user.(*conf.User), host.Group) {
			hosts = append(hosts, host.Host)
		}
	}

	var lastHost string
	var lastDest string
	var lastPort string
	go func() {
		for {
			select {
			case <-closed:
				log.Printf("close ws://%v\n", ws.RemoteAddr().String())
				if err := video.Send("STOP", lastHost, "", lastDest, lastPort); err != nil {
					log.Printf("send STOP error: %v\n", err)
				}
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
			log.Printf("resolve VideoSendRequest error: %v, message: %v\n", err, string(message))
			mu.Lock()
			util.WriteNil(ws, -201, "invalid request")
			mu.Unlock()
			continue
		}
		port := strconv.FormatInt(int64(v.Port), 10)
		lastHost = v.Host
		lastDest = v.Dest
		lastPort = port
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
