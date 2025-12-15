// Author: João Pinto
// Date: 2025-12-15
// Purpose: teste unitário para o tratamento de conteúdo HTML
// License: MIT

package minifier

import (
    "strings"
    "testing"
)

func TestHTMLPreserveAttributesInStyleScript(t *testing.T) {
    input := `<html><head><style media="screen"> body { color: red; } </style></head><body><script defer>const x = 1; // cmt
</script></body></html>`
    out := MinifyHTML(input, DefaultOptions())
    if out == "" { t.Fatal("output vazio") }
    if !strings.Contains(out, `<style media="screen">`) { t.Errorf("faltam atributos em style: %s", out) }
    if !strings.Contains(out, `<script defer>`) { t.Errorf("faltam atributos em script: %s", out) }
}

func TestHTMLPreservePreCode(t *testing.T) {
    input := `<pre>   A   B   C  
</pre><code>   X   Y   Z   </code>`
    out := MinifyHTML(input, DefaultOptions())
    if !strings.Contains(out, `<pre>   A   B   C  </pre>`) { t.Errorf("<pre> alterado: %s", out) }
    if !strings.Contains(out, `<code>   X   Y   Z   </code>`) { t.Errorf("<code> alterado: %s", out) }
}
