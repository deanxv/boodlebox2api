// @title KILO-AI-2API
// @version 1.0.0
// @description KILO-AI-2API
// @BasePath
package main

import (
	"boodlebox2api/check"
	"boodlebox2api/common"
	"boodlebox2api/common/config"
	logger "boodlebox2api/common/loggger"
	"boodlebox2api/middleware"
	"boodlebox2api/model"
	"boodlebox2api/router"
	"fmt"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

//var buildFS embed.FS

func main() {
	logger.SetupLogger()
	logger.SysLog(fmt.Sprintf("boodlebox2api %s starting...", common.Version))

	check.CheckEnvVariable()

	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	var err error

	model.InitTokenEncoders()
	config.InitSGCookies()

	server := gin.New()
	server.Use(gin.Recovery())
	server.Use(middleware.RequestId())
	middleware.SetUpLogger(server)

	// 设置API路由
	router.SetApiRouter(server)
	// 设置前端路由
	//router.SetWebRouter(server, buildFS)

	var port = os.Getenv("PORT")
	if port == "" {
		port = strconv.Itoa(*common.Port)
	}

	if config.DebugEnabled {
		logger.SysLog("running in DEBUG mode.")
	}

	logger.SysLog("boodlebox2api start success. enjoy it! ^_^\n")

	err = server.Run(":" + port)

	if err != nil {
		logger.FatalLog("failed to start HTTP server: " + err.Error())
	}
}
