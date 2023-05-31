package compiler

//go:generate go run github.com/evanw/esbuild/cmd/esbuild compiler_entry.ts --format=iife --global-name=__svelte__ --bundle --platform=node --inject:shimssr.ts --external:url --outfile=generated/compiler.js --log-level=warning
