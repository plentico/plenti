import { compile as compileSvelte } from "svelte/compiler"

type Input = {
  code: string
  path: string
  target: "ssr" | "dom"
  dev: boolean
  css: boolean
  name: string
}

// Capitalized for Go
type Output =
  | {
      JS: string
      CSS: string
    }
  | {
      Error: {
        Path: string
        Name: string
        Message: string
        Stack?: string
      }
    }

// Compile svelte code
export function compile(input: Input): string {
  const { code, path, target, dev, css, name } = input
  const svelte = compileSvelte(code, {
    name: name,
    filename: path,
    generate: target,
    hydratable: true,
    format: "esm",
    dev: dev,
    css: css,
  })
  return JSON.stringify({
    CSS: svelte.css.code,
    JS: svelte.js.code,
  } as Output)
}
