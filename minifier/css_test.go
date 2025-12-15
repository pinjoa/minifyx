// Author: João Pinto
// Date: 2025-12-15
// Purpose: teste unitário para o tratamento de conteúdo CSS
// License: MIT

package minifier

import "testing"

func TestCSSBasic(t *testing.T) {
    in := `/* cmt */ body { color: red ; margin : 0 ; }`
    out := MinifyCSS(in)
    if out != "body{color:red;margin:0;}" {
        t.Errorf("esperado 'body{color:red;margin:0;}', obtido '%s'", out)
    }
}
