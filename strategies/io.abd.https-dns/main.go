package main

import (
	"fmt"

	"github.com/appc/abd/schema"
	"github.com/appc/abd/strategies"
)

const StrategyIoAbdHttpDnsName = "io.abd.http-dns"

type StrategyIoAbdHttpDns struct {
	schema.ABDMetadataFetchStrategyConfiguration
}

func main() {
	strategies.PluginMain(getHttpDnsMetadata)
}

func getHttpDnsMetadata(args *strategies.PluginConf) error {
	fmt.Println("httpdns plugin")
	fmt.Println("called with identifier:")
	fmt.Println(args.Identifier)
	fmt.Println("got config:")
	fmt.Println(args.Configuration)
	return nil
}
