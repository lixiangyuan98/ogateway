package model

import (
	"github.com/lixiangyuan98/ogateway/control/video"
)

type WebsocketResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type VideoSendRequest struct {
	Method string `json:"method" binding:"required"`
	Host   string `json:"host" binding:"required"`
	Src    string `json:"src" binding:"required"`
	Dest   string `json:"dest" binding:"required"`
	Port   uint16 `json:"port" binding:"required"`
}

type VideoSendResponse struct {
	Servers []*video.VideoServer `json:"servers"`
}
