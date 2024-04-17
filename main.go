package main

import (
	"encoding/json"
	"github.com/cloudwego/hertz/cmd/hz/meta"
	"github.com/cloudwego/thriftgo/plugin"
	_ "github.com/cloudwego/thriftgo/plugin"
	"io"
	"log"
	"os"
)

var debug = false // TODO: add flags somehow
var filename = "openapi.yaml"

func main() {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Println("Failed to get input:", err.Error())
		os.Exit(1)
	}
	req, err := plugin.UnmarshalRequest(data)
	if err != nil {
		log.Println("Failed to unmarshal request:", err.Error())
		os.Exit(1)
	}
	if debug {
		debugPrint(req)
	}

	os.Exit(exit(NewOpenAPIv3Generator(req).Run()))
}

func debugPrint(req *plugin.Request) {
	reqJson, err := json.Marshal(req)
	if err != nil {
		return // ignore error
	}
	err = os.WriteFile("plugin_request_sample.json", reqJson, 0644)
	if err != nil {
		return // ignore error
	}
}

func exit(res *plugin.Response) int {
	data, err := plugin.MarshalResponse(res)
	if err != nil {
		log.Println("Failed to marshal response:", err.Error())
		return meta.PluginError
	}
	_, err = os.Stdout.Write(data)
	if err != nil {
		log.Println("Error at writing response out:", err.Error())
		return meta.PluginError
	}
	return 0
}
