package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/appc/abd/Godeps/_workspace/src/github.com/spf13/cobra"

	"github.com/appc/abd/schema"
)

const (
	defaultConfigDir = "/usr/lib/abd/sources.list.d/"

	exampleLocalStrategyConfiguration = `
	{
		"prefix": "*",
		"strategy": "io.abd.local",
		"storage-path": "/var/abd"
	}
	`
)

var (
	configDir = defaultConfigDir
	abdCmd    = &cobra.Command{
		Use:   "abd",
		Short: "abd - the appc Binary Discovery",
		Long:  `abd is a framework for resolving human-readable strings to downloadable URIs`,
	}
	discoverCmd = &cobra.Command{
		Use:   "discover",
		Short: "discover an artefact using ABD",
		Run:   discoverFunc,
	}
	mirrorsCmd = &cobra.Command{
		Use:   "mirrors",
		Short: "discover the mirrors for an artefact using ABD",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("not implemented")
		},
	}
	fetchCmd = &cobra.Command{
		Use:   "fetch",
		Short: "discover and fetch an artefact using ABD",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("not implemented")
		},
	}
)

func init() {
	abdCmd.PersistentFlags().StringVarP(&configDir, "config-dir", "", "", "configuration directory for abd")
	abdCmd.AddCommand(discoverCmd, mirrorsCmd, fetchCmd)
}

func main() {
	abdCmd.Execute()
}

func discoverFunc(cmd *cobra.Command, args []string) {
	// abd discover identifier,label1=value1,label2=value2,...
	if len(args) < 1 {
		fmt.Println("no identifier given")
		os.Exit(1)
	}

	// TODO(jonboulle): define proper string -> identifier/labels parser
	parts := strings.SplitN(args[0], ",", 2)
	identifier := parts[0]
	if identifier == "" {
		fmt.Println("no identifier given")
		os.Exit(1)
	}
	// TODO(jonboulle): define proper arg interface for plugins for labels
	// for now, just label1=value1 label2=value2, etc
	var lbls []string
	if len(parts) > 1 {
		v, err := url.ParseQuery(strings.Replace(parts[1], ",", "&", -1))
		if err != nil {
			fmt.Println("error parsing labels:", err.Error())
			os.Exit(1)
		}
		for key, val := range v {
			if len(val) > 1 {
				fmt.Println("label with multiple values:", key)
				os.Exit(1)
			}
			lbls = append(lbls, fmt.Sprintf("%s=%s", key, val[0]))
		}
	}

	cfgs, err := getConfigs()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	c := cfgs[0]

	// prefix matching the config
	if !strings.HasPrefix(identifier, c.Prefix) &&
		c.Prefix != "*" {
		fmt.Println("no matching prefix found")
		os.Exit(1)
	}

	// dumb: just look for executable with same name as strategy
	path, err := exec.LookPath(c.Strategy)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	cfg, err := json.Marshal(c)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	run := exec.Cmd{
		Path:  path,
		Args:  append([]string{identifier}, lbls...),
		Stdin: bytes.NewBuffer(cfg),
	}
	out, err := run.CombinedOutput()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println(string(out))
}

// getConfigs will loop through the config directory and return the parsed
// version of all of the configs.
func getConfigs() ([]schema.ABDMetadataFetchStrategyConfiguration, error) {
	fi, err := ioutil.ReadDir(configDir)
	if err != nil {
		return nil, err
	}

	cfgs := make([]schema.ABDMetadataFetchStrategyConfiguration, len(fi))

	for i := range fi {
		f, err := os.Open(filepath.Join(configDir, fi[i].Name()))
		if err != nil {
			return nil, err
		}

		b, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}
		f.Close()

		var c schema.ABDMetadataFetchStrategyConfiguration
		err = json.Unmarshal(b, &c)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		cfgs[i] = c
	}

	return cfgs, nil
}
