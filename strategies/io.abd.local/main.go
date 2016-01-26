package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	"github.com/appc/abd/schema"
	"github.com/appc/abd/strategies"
)

const StrategyIoAbdLocalName = "io.abd.local"

type StrategyIoAbdLocal struct {
	schema.ABDMetadataFetchStrategyConfiguration
	StoragePath string `json:"storage-path"`
}

func main() {
	strategies.PluginMain(getLocalMetadata)
}

func getLocalMetadata(args *strategies.PluginConf) error {
	fmt.Println("io.abd.local plugin")
	fmt.Println("  called with identifier: " + args.Identifier)
	fmt.Printf("  called with labels: %v\n", args.Labels)
	fmt.Println("  called with config: " + args.Configuration)
	var conf StrategyIoAbdLocal
	err := json.Unmarshal([]byte(args.Configuration), &conf)
	if err != nil {
		panic("bad configuration supplied")
	}
	fmt.Println("ABD metadata blob:")
	t.Execute(os.Stdout, struct{ Identifier, StoragePath string }{args.Identifier, conf.StoragePath})

	return nil
}

var t = template.Must(template.New("").Parse(strategyTemplate))

const strategyTemplate = `
	{
		"name": "{{.Identifier}}",
		"mirrors": [
			 "{{.StoragePath}}/{{.Identifier}}",
		]
	}
	`
