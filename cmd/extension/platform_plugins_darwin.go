package main

import (
	"os"

	"github.com/go-kit/kit/log"
	osquery "github.com/kolide/osquery-go"

	"github.com/kolide/go-extension-tutorial/pkg/mdm"
	"github.com/kolide/go-extension-tutorial/pkg/spotlight"
)

func registerPlatformPlugins(server *osquery.ExtensionManagerServer) {
	logger := log.NewLogfmtLogger(os.Stderr)
	mdmPlugin := mdm.MDMInfo(logger)
	server.RegisterPlugin(mdmPlugin)

	// register the macOS spotlight plugin.
	server.RegisterPlugin(spotlight.New())
}
