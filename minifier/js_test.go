// Author: João Pinto
// Date: 2025-12-15
// Purpose: teste unitário para o tratamento de conteúdo JS
// License: MIT

package minifier

import (
    "strings"
    "testing"
)

func TestJSCommentsAndStrings(t *testing.T) {
    in := `const a = "http://example.com"; // linha
/* bloco */ const b = '\' ;`
    out := MinifyJS(in)
    if out == "" { t.Fatal("output vazio") }
    if !strings.Contains(out, `"http://example.com"`) { t.Errorf("URL corrompida: %s", out) }
}
