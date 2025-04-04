//go:generate bash -c "if [ ! -f go.mod ]; then echo 'Initializing go.mod...'; go mod init .containifyci; else echo 'go.mod already exists. Skipping initialization.'; fi"
//go:generate go get github.com/containifyci/engine-ci/protos2
//go:generate go get github.com/containifyci/engine-ci/client
//go:generate go mod tidy

package main

import (
	"os"

	"github.com/containifyci/engine-ci/client/pkg/build"
)

func main() {
	_ = os.Chdir("../")
	opts := build.NewGoServiceBuild("go-mask")
	opts.Verbose = false
	opts.Image = ""
	opts.File = "main.go"
	opts.Properties = map[string]*build.ListValue{
		"coverage_mode": build.NewList("binary"),
	}

	opts2 := build.NewGoServiceBuild("go-mask-main")
	opts2.Verbose = false
	opts2.Image = ""
	opts2.File = "main.go"
	opts2.Properties = map[string]*build.ListValue{
		"tags":            build.NewList("testrunmain"),
		"nocoverage":      build.NewList("true"),
		"goreleaser":      build.NewList("true"),
		"goreleaser_envs": build.NewList("MACOS_SIGN_PASSWORD", "MACOS_SIGN_P12", "MACOS_NOTARY_ISSUER_ID", "MACOS_NOTARY_KEY_ID", "MACOS_NOTARY_KEY"),
	}
	build.Serve(opts, opts2)
}
