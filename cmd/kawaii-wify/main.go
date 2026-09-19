package main

import "github.com/obliviousorion/kawaii-wify/internal/cli"

var Version = "dev"

func main() {
	cli.Execute(Version)
}
