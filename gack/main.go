package main

import (
	"fmt"
	"os"

	"github.com/abates/gack"
	"github.com/abates/gack/build"
	"github.com/abates/gack/generator"
	"github.com/abates/gack/pkg"
)

var mux *gack.Mux

func usage(messages ...string) {
	for _, message := range messages {
		fmt.Fprintf(os.Stderr, "%s\n", message)
	}
	fmt.Fprintf(os.Stderr, "Usage: %s [target]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Available targets are:\n")
	for _, target := range mux.TargetNames() {
		fmt.Fprintf(os.Stderr, "\t%v\n", target)
	}
	os.Exit(1)
}

func main() {
	var err error
	mux, err = gack.NewMux()
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: %v\n", err)
	}
	build.Register(mux)
	generator.Register(mux)
	pkg.Register(mux)

	if len(os.Args) == 1 {
		usage()
	}

	err = mux.Execute(os.Args[1])

	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}
