package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/DevLabFoundry/configmanager-plugin-awsparamstr/impl"
	"github.com/DevLabFoundry/configmanager/v3/config"
	"github.com/DevLabFoundry/configmanager/v3/tokenstore"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

type implIface interface {
	Value(token string, metadata []byte) (string, error)
}
type TokenStorePlugin struct {
	impl implIface
}

func (ts TokenStorePlugin) Value(key string, metadata []byte) (string, error) {
	return ts.impl.Value(key, metadata)
}

var (
	Version  string = "0.0.1"
	Revision string = "1111aaaa"
)

func main() {

	versionFlag := flag.Bool("version", false, "plugin version")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("Version: %s-%s\n", Version, Revision)
		os.Exit(0)
	}

	// log set up
	log := hclog.New(hclog.DefaultOptions)
	log.SetLevel(hclog.LevelFromString("error"))

	if val, ok := os.LookupEnv(config.CONFIGMANAGER_LOG); ok && len(val) > 0 {
		if logLevel := hclog.LevelFromString(val); logLevel != hclog.NoLevel {
			log.SetLevel(logLevel)
		}
	}

	// initialize the implementation
	i, err := impl.NewParamStore(context.Background(), log)
	if err != nil {
		log.Error("error", err)
		os.Exit(1)
	}

	// initializing the service
	ts := TokenStorePlugin{impl: i}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: tokenstore.Handshake,
		Plugins: map[string]plugin.Plugin{
			"configmanager_token_store": &tokenstore.GRPCPlugin{Impl: ts},
		},
		VersionedPlugins: map[int]plugin.PluginSet{
			1: {
				"configmanager_token_store": &tokenstore.GRPCPlugin{Impl: ts},
			},
		},
		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
