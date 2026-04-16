package app

import (
	"wa_chat_service/config"

	"go.uber.org/zap"
)

func Run(config *config.Config, zslog *zap.SugaredLogger) {
	zslog.Infof("Starting %s, version %s, in %s mode", config.App.Name, config.App.Version, config.App.Environment)
	servers := NewDefaultWiring(zslog, config)
	servers.startServers()
	servers.waitForShutdown(zslog.Desugar())
	zslog.Info("Server exiting")
}
