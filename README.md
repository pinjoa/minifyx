# MinifyX

MinifyX é uma ferramenta e biblioteca escrita em **Go**, criada para minificar **HTML, CSS, JavaScript, JSON e XML**, preservando atributos em `<style>` e `<script>` e evitando minificação dentro de `<pre>` e `<code>`.

O objetivo é oferecer uma solução simples, rápida e segura para integrar em pipelines, scripts ou aplicações Go.

---

## ✨ Funcionalidades

- Minificação de:
  - HTML
  - CSS
  - JavaScript
  - JSON
  - XML
- Preserva atributos inline em `<style>` e `<script>`
- Evita minificação dentro de `<pre>` e `<code>`
- Disponível como **CLI** e como **biblioteca Go**
- Sem dependências externas pesadas
- Simples de integrar em automações

---

## 📦 Instalação

### Instalar via `go install`

```bash
go install github.com/pinjoa/minifyx/cmd/minifyx@latest
```

O binário ficará disponível em:

```bash
$GOPATH/bin/minifyx
```

---

## 🚀 Utilização (CLI)

### Minificar um ficheiro HTML

```bash
minifyx index.html
```

### Forçar tipo de ficheiro

```bash
minifyx -type js snippet.txt
```

---

## 🧩 Utilização como Biblioteca Go

### Exemplo básico

```go
package main

import (
    "fmt"
    "os"

    "github.com/pinjoa/minifyx/minifier"
)

func main() {
    input, _ := os.ReadFile("index.html")

    result, err := minifier.MinifyHTML(string(input))
    if err != nil {
        panic(err)
    }

    fmt.Println(result)
}
```

### Minificar CSS

```go
css := `
body {
    color: red;
    margin: 0px;
}
`

min, err := minifier.MinifyCSS(css)
if err != nil {
    panic(err)
}

fmt.Println(min)
```

### Minificar JSON

```go
json := `{
    "name": "João",
    "age": 30
}`

min, _ := minifier.MinifyJSON(json)
fmt.Println(min)
```

---

```
minifyx [opções] <ficheiros...>
```

## ⚙️ Opções da CLI

| Opção                   | Descrição                                                       |
| ----------------------- | --------------------------------------------------------------- |
| `-o`                    | Define ficheiro ou diretório de saída                           |
| `-type`                 | Forçar tipo: html,css,js,json,xml                               |
| `-no-html-json`         | Não minificar JSON em `<script type="application/*json">`       |
|                         | e atributos data-json                                           |
| `-no-html-templates`    | Não minificar HTML dentro de `<template>` e scripts de template |
| `-no-html-whitespace`   | Não colapsar espaços em HTML (texto fora de blocos especiais)   |
| `-no-xml-whitespace`    | Não colapsar espaços/indentação em XML                          |
| `-parallel`             | Número de goroutines em paralelo (default nº de cpu)            |
| `-preserve-precode`     | Preservar conteúdo especial em `<pre>/<code>`                   |
|                         | (e tratar `<code>` como bloco) (default true)                   |
| `-remove-html-comments` | Remover comentários HTML (default true)                         |
| `-remove-xml-comments`  | Remover comentários XML (default true)                          |
| `-stdin`                | Ler de stdin                                                    |
| `-stdout`               | Escrever para stdout                                            |

---

## 📚 Exemplos de Workflow

### Automatizar minificação com Bash

```bash
for f in *.html; do
    minifyx -o "dist/" "$f"
done
```

### Integrar num build Go

```go
func buildAssets() {
    files := []string{"index.html", "style.css", "app.js"}

    for _, f := range files {
        data, _ := os.ReadFile(f)
        min, _ := minifier.Auto(f, string(data))
        os.WriteFile("dist/"+f, []byte(min), 0644)
    }
}
```

---

## 📄 Licença

Licenciado sob a **MIT License**.
