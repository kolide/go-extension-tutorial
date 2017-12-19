package main

import (
	osquery "github.com/kolide/osquery-go"

	"github.com/kolide/go-extension-tutorial/pkg/spotlight"
)

func registerPlatformPlugins(server *osquery.ExtensionManagerServer) {
	// register the macOS spotlight plugin.
	server.RegisterPlugin(spotlight.New())
}
