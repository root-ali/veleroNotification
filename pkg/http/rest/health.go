package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/root-ali/velero-reporter/pkg/health"
)

func Ready(hs health.HealthService) gin.HandlerFunc {
	return func(context *gin.Context) {
		err := hs.Ready()
		if err != nil {
			context.JSON(500, gin.H{
				"message": "server error",
			})
		} else {
			context.JSON(200, gin.H{
				"message": "OK",
			})
		}
	}
}

func Healthy(hs health.HealthService) gin.HandlerFunc {
	return func(context *gin.Context) {
		err := hs.Healthy()
		if err != nil {
			context.JSON(500, gin.H{
				"message": "server error",
			})
		} else {
			context.JSON(200, gin.H{
				"message": "OK",
			})
		}
	}
}
