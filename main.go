package main

//go:generate go run generators/defaults.go

import "plenti/cmd"

func main() {
	cmd.Execute()
}
