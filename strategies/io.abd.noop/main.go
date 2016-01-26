package main

import (
	"fmt"

	"github.com/appc/abd/schema"
	"github.com/appc/abd/strategies"
)

const StrategyIoAbdNoopName = "io.abd.noop"

type StrategyIoAbdNoop struct {
	schema.ABDMetadataFetchStrategyConfiguration
}

func main() {
	strategies.PluginMain(getNoopMetadata)
}

func getNoopMetadata(args *strategies.PluginConf) error {
	fmt.Println("{}")
	return nil
}
