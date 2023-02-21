// Command protoc-gen-grpc-gateway is a plugin for Google protocol buffer
// compiler to generate a reverse-proxy, which converts incoming RESTful
// HTTP/1 requests gRPC invocation.
// You rarely need to run this program directly. Instead, put this program
// into your $PATH with a name "protoc-gen-grpc-gateway" and run
//
//	protoc --grpc-gateway_out=output_directory path/to/input.proto
//
// See README.md for more details.
package main

import (
	"flag"
	"github.com/Mikaelemmmm/protobuf/internal/descriptor"
	"github.com/Mikaelemmmm/protobuf/plugin/protoc-gen-zeroapi/internal/gozero"
	"google.golang.org/protobuf/compiler/protogen"
)

var (
	name     = flag.String("name", "Gateway", "gen code app name")
	out      = flag.String("out", "gateway", "gen code dir path")
	template = flag.String("template", "template", "gen code template dir path")
)

func main() {
	flag.Parse()

	protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}.Run(func(gen *protogen.Plugin) error {
		reg := descriptor.NewRegistry()
		if err := reg.LoadFromPlugin(gen); err != nil {
			return err
		}
		targets := make([]*descriptor.File, 0, len(gen.Request.FileToGenerate))
		for _, target := range gen.Request.FileToGenerate {
			f, err := reg.LookupFile(target)
			if err != nil {
				return err
			}
			targets = append(targets, f)
		}

		generator := gozero.New(gozero.WithApp(*name), gozero.WithRootDir(*out), gozero.WithTemplateDir(*template))
		return generator.Generate(targets)
	})
}
