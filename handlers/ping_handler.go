package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type PingHandler struct{}

func NewPingHandler() *PingHandler {
	return &PingHandler{}
}

func (ph *PingHandler) Ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}
