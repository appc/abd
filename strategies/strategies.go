// Common strategies code, defining the plugin execution interface
package strategies

import (
	"io/ioutil"
	"os"
	"strings"
)

type PluginConf struct {
	Identifier    string
	Labels        map[string]string
	Configuration string
}

// PluginMain is the shared interface for all strategy plugins
// <plugin> <identifier> <lbl1=val1> <lbl2=val1> ...
func PluginMain(cmd func(_ *PluginConf) error) {
	if len(os.Args) < 1 {
		panic("not enough args!")
	}
	pc := PluginConf{
		Identifier: os.Args[0],
		Labels:     make(map[string]string),
	}

	var labels []string
	if len(os.Args) > 1 {
		labels = os.Args[1:]
	}
	for _, lbl := range labels {
		kv := strings.SplitN(lbl, "=", 2)
		if len(kv) != 2 {
			panic("badly formed label!")
		}
		if _, ok := pc.Labels[kv[0]]; ok {
			panic("label set twice!")
		}
		pc.Labels[kv[0]] = kv[1]
	}
	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic("error reading config from stdin!")
	}
	pc.Configuration = string(b)
	if err = cmd(&pc); err != nil {
		panic("plugin execution failed!" + err.Error())
	}
}
