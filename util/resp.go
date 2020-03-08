package util

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func WriteJson(ws *websocket.Conn, code int, message string, data interface{}) {
	resp := map[string]interface{}{
		"code":    code,
		"message": message,
		"data":    data,
	}
	if err := ws.WriteJSON(resp); err != nil {
		log.Printf("write message to ws://%v error: %v\n", ws.RemoteAddr().String(), err)
	}
}

func WriteNil(ws *websocket.Conn, code int, message string) {
	resp := map[string]interface{}{
		"code":    code,
		"message": message,
	}
	if err := ws.WriteJSON(resp); err != nil {
		log.Printf("write message to ws://%v error: %v\n", ws.RemoteAddr().String(), err)
	}
}

func RetJson(ctx *gin.Context, code int, message string, data interface{}) {
	ctx.JSON(200, gin.H{
		"code":    code,
		"message": message,
		"data":    data,
	})
}

func RetErr(ctx *gin.Context, code int, message string) {
	ctx.JSON(200, gin.H{
		"code":    code,
		"message": message,
	})
}
