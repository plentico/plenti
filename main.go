package main

//go:generate go run generators/main.go

import "plenti/cmd"

func main() {
	cmd.Execute()
}
