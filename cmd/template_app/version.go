package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

var (
	buildstamp  = "not set"
	buildnumber = "not set"
	githash     = "not set"
)

func showVersion() {
	fmt.Println(appname)
	fmt.Println(" Git hash: ", githash)
	fmt.Println(" Build time: ", buildstamp)
	fmt.Println(" Build number: ", buildnumber)
	fmt.Println(" Platform:", runtime.GOOS, "/", runtime.GOARCH)
	if bi, ok := debug.ReadBuildInfo(); ok {
		fmt.Println(" Go version:", bi.GoVersion)
		fmt.Println(" Main module:", bi.Main.Path)
		fmt.Println(" Modules:", bi.Main.Version)
	}
}
