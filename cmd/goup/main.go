package main

import (
	"github.com/mirkobrombin/goup/internal/cli"
	"github.com/mirkobrombin/goup/internal/plugin"
	"github.com/mirkobrombin/goup/plugins"
)

func main() {
	pluginManager := plugin.GetPluginManagerInstance()

	// Register your plugins here:
	pluginManager.Register(&plugins.CustomHeaderPlugin{})
	pluginManager.Register(&plugins.PHPPlugin{})

	cli.Execute()
}
