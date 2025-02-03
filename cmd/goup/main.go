package main

import (
	"github.com/mirkobrombin/goup/internal/cli"
	"github.com/mirkobrombin/goup/internal/plugin"
	"github.com/mirkobrombin/goup/plugins"
)

func main() {
	pluginManager := plugin.NewPluginManager()
	plugin.SetDefaultPluginManager(pluginManager)

	// Register your plugins here:
	pluginManager.Register(&plugins.CustomHeaderPlugin{})
	pluginManager.Register(&plugins.PHPPlugin{})
	pluginManager.Register(&plugins.AuthPlugin{})
	pluginManager.Register(&plugins.NodeJSPlugin{})

	cli.Execute()
}
