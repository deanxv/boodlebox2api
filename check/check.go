package check

import (
	"boodlebox2api/common/config"
	logger "boodlebox2api/common/loggger"
)

func CheckEnvVariable() {
	logger.SysLog("environment variable checking...")

	if config.BBCookie == "" {
		logger.FatalLog("环境变量 BB_COOKIE 未设置")
	}
	if config.UserId == "" {
		logger.FatalLog("环境变量 USER_ID 未设置")
	}

	logger.SysLog("environment variable check passed.")
}
