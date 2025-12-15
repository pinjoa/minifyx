// Author: João Pinto
// Date: 2025-12-15
// Purpose: Módulo para testar a integração de diferentes blocos dentro do HTML
// License: MIT

package minifier

import "testing"

func TestHTMLFull(t *testing.T) {
    input := `<!DOCTYPE html>
<html>
<head>
<style media="screen">
/* cmt */ body { color: red ; }
</style>
</head>
<body>
<script defer>
// linha
const url = "http://x"; /* bloco */
</script>
<pre>  X   Y   Z  </pre>
</body>
</html>`
    out := MinifyHTML(input, DefaultOptions())
    if out == "" { t.Fatal("output vazio") }
    if !contains(out, `<style media="screen">`) { t.Error("atributo media perdido em <style>") }
    if !contains(out, `<script defer>`) { t.Error("atributo defer perdido em <script>") }
    if !contains(out, `<pre>  X   Y   Z  </pre>`) { t.Error("conteúdo <pre> alterado") }
}

func contains(s, sub string) bool { for i := 0; i+len(sub) <= len(s); i++ { if s[i:i+len(sub)] == sub { return true } } ; return false }
