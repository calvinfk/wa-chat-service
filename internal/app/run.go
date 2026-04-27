package app

import (
	"wa_chat_service/config"

	"go.uber.org/zap"
)

func Run(config *config.Config, zsLog *zap.SugaredLogger) {
	zsLog.Infof("Starting %s, version %s, in %s mode", config.App.Name, config.App.Version, config.App.Environment)
	servers := NewDefaultWiring(zsLog, config)
	servers.startServers()
	servers.waitForShutdown(zsLog.Desugar())
}
