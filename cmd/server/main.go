package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/root-ali/velero-reporter/pkg/Config"
	"github.com/root-ali/velero-reporter/pkg/health"
	vr_http "github.com/root-ali/velero-reporter/pkg/http"
	"github.com/root-ali/velero-reporter/pkg/kubernetes"
	"github.com/root-ali/velero-reporter/pkg/notifier/mattermost"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type veleroReporter struct {
	config    Config.VeleroReportorConfig
	HS        health.HealthService
	KC        *kubernetes.KubernetesClient
	HT        *gin.Engine
	MC        *mattermost.MattermostClient
	webServer *http.Server
	l         *zap.SugaredLogger
}

func main() {

	vr := newVeleroReporter()
	vr.loadConfig()
	vr.loadMattermostClient()
	vr.loadKubeConfig()
	vr.loadHealthService()
	vr.runServer()
	vr.awaitTermination()

}

func (vr *veleroReporter) loadMattermostClient() {
	mc := mattermost.NewMattermostClient(vr.config.MattermostUrl, vr.config.MattermostToken, "testing", 10*time.Second, vr.l)
	vr.MC = mc
}

func newVeleroReporter() *veleroReporter {
	vr := &veleroReporter{}
	vr.loadConfig()
	config := zap.NewProductionConfig()
	if vr.config.LogLevel == "info" || vr.config.LogLevel == "" {
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	if vr.config.LogLevel == "debug" {
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	sugar := logger.Sugar()
	vr.l = sugar
	vr.l.Info("velero reporter initialized", vr)
	return vr
}

func (vr *veleroReporter) loadConfig() {
	viper.AutomaticEnv()
	err := viper.BindEnv("mattermostUrl", "VELERO_REPORTER_MATTERMOST_URL")
	err = viper.BindEnv("mattermostToken", "VELERO_REPORTER_MATTERMOST_TOKEN")
	err = viper.BindEnv("kubeConfigType", "VELERO_REPORTER_KUBECONFIG_TYPE")
	err = viper.BindEnv("kubeConfigPath", "VELERO_REPORTER_KUBECONFIG_PATH")
	err = viper.BindEnv("logLevel", "LOG_LEVEL", "LOG_LEVEL")
	err = viper.BindEnv("httpHost", "VELERO_REPORTER_HTTP_HOST")
	err = viper.BindEnv("httpPort", "VELERO_REPORTER_HTTP_PORT")
	if err != nil {
		fmt.Println("error is ", err)
	}
	vr.config.MattermostUrl = viper.GetString("mattermostUrl")
	vr.config.MattermostToken = viper.GetString("mattermostToken")
	vr.config.KubeConfigType = viper.GetString("kubeConfigType")
	vr.config.KubeConfigPath = viper.GetString("kubeConfigPath")
	vr.config.HttpHost = viper.GetString("httpHost")
	vr.config.HttpPort = viper.GetString("httpPort")
	vr.config.LogLevel = viper.GetString("logLevel")
}

func (vr *veleroReporter) loadKubeConfig() {
	vr.KC = kubernetes.NewKubernetesClient(vr.l, vr.config.KubeConfigPath, vr.MC)
}

func (vr *veleroReporter) loadHealthService() {
	vr.HS = health.NewHealthService(vr.KC, vr.l)
}

func (vr *veleroReporter) runServer() {
	httpHandler := vr_http.NewHttpService(vr.HS, vr.l)
	vr.HT = httpHandler.Handler()

	vr.webServer = &http.Server{
		Addr:    vr.config.HttpHost + ":" + vr.config.HttpPort,
		Handler: vr.HT,
	}
	go func() {
		if err := vr.webServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			vr.l.Fatalf("listen: %s\n", err)
		}
	}()
}

func (vr *veleroReporter) awaitTermination() {
	receiver := make(chan os.Signal)
	signal.Notify(receiver, os.Interrupt, os.Kill)

	<-receiver
	vr.l.Debug("Received interrupt signal!")

	vr.stop()

}

func (vr *veleroReporter) stop() {

	vr.l.Info("Shutting down server...")
	vr.KC.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := vr.webServer.Shutdown(ctx); err != nil {
		vr.l.Fatalf("Server forced to shutdown: %v", err)
	}
	vr.l.Info("Server exiting")
}
