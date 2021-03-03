package main

//go:generate go run generators/main.go

import "github.com/plentico/plenti/cmd"

func main() {
	cmd.Execute()
}
