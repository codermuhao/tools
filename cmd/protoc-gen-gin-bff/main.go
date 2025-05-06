package main

import (
	"flag"
	"fmt"

	"github.com/codermuhao/tools/cmd/protoc-gen-gin-bff/internal/generate"
	"google.golang.org/protobuf/compiler/protogen"
)

var showVersion = flag.Bool("version", false, "print the version and exit")

func main() {
	flag.Parse()
	if *showVersion {
		fmt.Printf("protoc-gen-error %v\n", release)
		return
	}
	var flags flag.FlagSet
	protogen.Options{ParamFunc: flags.Set}.Run(func(gen *protogen.Plugin) error {
		g := generate.NewGen(gen)
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			g.File(f, release)
		}
		return nil
	})
}
