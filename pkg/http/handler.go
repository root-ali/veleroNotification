package http

import (
	_ "github.com/danielkov/gin-helmet"
	helmet "github.com/danielkov/gin-helmet"
	_ "github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/penglongli/gin-metrics/ginmetrics"
	"github.com/root-ali/velero-reporter/pkg/http/rest"
)

func (ht *HttpHandler) Handler() *gin.Engine {
	router := gin.Default()
	if gin.Mode() != "production" && gin.Mode() != "test" {
		gin.SetMode(gin.DebugMode)
	} else if gin.Mode() == "test" {
		gin.SetMode(gin.TestMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	m := ginmetrics.GetMonitor()

	m.SetMetricPath("/metrics")
	m.SetSlowTime(10)
	m.SetDuration([]float64{0.1, 0.3, 1.2, 5, 10})
	m.Use(router)
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			ht.l.Infow("gin_access_log",
				"ClientIP", param.ClientIP,
				"Timestamp", param.TimeStamp,
				"Method", param.Method,
				"Path", param.Path,
				"Protocol", param.Request.Proto,
				"StatusCode", param.StatusCode,
				"Latency", param.Latency,
				"UserAgent", param.Request.UserAgent(),
				"ErrorMessage", param.ErrorMessage,
			)
			return ""
		},
	}))
	router.Use(gin.Recovery())
	router.Use(helmet.Default())
	router.SetTrustedProxies(nil)
	router.GET("/ready", rest.Ready(ht.hs))
	router.GET("/healthy", rest.Healthy(ht.hs))
	return router
}
