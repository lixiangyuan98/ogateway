package main

import (
	"crypto/md5"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/lixiangyuan98/ogateway/conf"
	"github.com/lixiangyuan98/ogateway/control/video"
	"github.com/lixiangyuan98/ogateway/handler"
	"github.com/lixiangyuan98/ogateway/util"
)

func main() {
	r := gin.Default()
	r.Use(loginMiddleware)
	r.POST("/send", handler.Send)
	r.GET("/ws", handler.ConnectVideoServer)
	r.Run(fmt.Sprintf("%v:%v", conf.GlobalConf.IP, conf.GlobalConf.Port))
}

func init() {
	conf.Init()
	// 启动和视频采集端的通信
	go video.Init()
}

func loginMiddleware(ctx *gin.Context) {
	username := ctx.Request.Header.Get("username")
	password := fmt.Sprintf("%x", md5.Sum([]byte(ctx.Request.Header.Get("password"))))
	user, ok := conf.UserConf[username]
	if !ok {
		ctx.Abort()
		util.RetErr(ctx, -100, "no such user")
		return
	}
	if user.Password != password {
		ctx.Abort()
		util.RetErr(ctx, -101, "wrong password")
		return
	}
	ctx.Set("user", user)
	ctx.Next()
}
