package port

import "github.com/gin-gonic/gin"

type GinMuxAttacher interface {
	AttachGinMux(*gin.Engine)
}
