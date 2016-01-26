package main

import (
	"fmt"

	"github.com/appc/abd/schema"
	"github.com/appc/abd/strategies"
)

const StrategyIoAbdNfsName = "io.abd.nfs"

type StrategyIoAbdNfs struct {
	schema.ABDMetadataFetchStrategyConfiguration
	FetchUri string `json:"fetch-uri"`
}

func main() {
	strategies.PluginMain(getNfsMetadata)
}

func getNfsMetadata(args *strategies.PluginConf) error {
	fmt.Println("nfs plugin")
	fmt.Println("called with identifier:")
	fmt.Println(args.Identifier)
	fmt.Println("got config:")
	fmt.Println(args.Configuration)
	return nil
}
