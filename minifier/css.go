// Author: João Pinto
// Date: 2025-12-15
// Purpose: MinifyCSS faz uma minificação conservadora de CSS
//          remove comentários de bloco fora de strings e aperta espaços supérfluos
// License: MIT

package minifier

import (
	"strings"
)

func MinifyCSS(input string) string {
    var out strings.Builder
    inBlockComment := false
    inString := false
    inURL := false
    urlParenDepth := 0
    stringQuote := byte(0)
    lastSpaceOutsideString := false

    b := []byte(input)

    for i := 0; i < len(b); i++ {
        c := b[i]
        var next byte
        if i+1 < len(b) {
            next = b[i+1]
        }

        // Detetar contexto url(...) (case-insensitive) fora de strings/comentários
        if !inString && !inBlockComment {
            if !inURL && (c == 'u' || c == 'U') && i+3 < len(b) {
                u := b[i : i+3]
                if (u[0] == 'u' || u[0] == 'U') &&
                    (u[1] == 'r' || u[1] == 'R') &&
                    (u[2] == 'l' || u[2] == 'L') &&
                    b[i+3] == '(' {
                    inURL = true
                    urlParenDepth = 0
                }
            }

            if inURL {
                switch c {
                case '(':
                    urlParenDepth++
                case ')':
                    if urlParenDepth > 0 {
                        urlParenDepth--
                    }
                    if urlParenDepth == 0 {
                        inURL = false
                    }
                }
            }
        }

        if inBlockComment {
            if c == '*' && next == '/' {
                inBlockComment = false
                i++
            }
            continue
        }

        // Início de comentário de bloco (mas preserva /*! ... */)
        if !inString && !inURL && c == '/' && next == '*' {
            if i+2 < len(b) && b[i+2] == '!' {
                // comentário do tipo /*! ... */ é mantido
                out.WriteByte(c)
                out.WriteByte(next)
                i++
                lastSpaceOutsideString = false
                continue
            }
            inBlockComment = true
            i++
            continue
        }

        // Strings
        if !inString && (c == '\'' || c == '"') {
            inString = true
            stringQuote = c
            out.WriteByte(c)
            lastSpaceOutsideString = false
            continue
        }

        if inString {
            out.WriteByte(c)
            if c == '\\' {
                if i+1 < len(b) {
                    out.WriteByte(b[i+1])
                    i++
                }
                continue
            }
            if c == stringQuote {
                inString = false
            }
            // não tocamos em lastSpaceOutsideString aqui
            continue
        }

        // Ignorar quebras de linha / tabs fora de strings
        if c == '\n' || c == '\r' || c == '\t' {
            continue
        }

        // Colapsar múltiplos espaços fora de strings
        if c == ' ' {
            if lastSpaceOutsideString {
                continue // já temos um espaço, não precisamos de outro
            }
            out.WriteByte(' ')
            lastSpaceOutsideString = true
            continue
        }

        // Qualquer outro caractere “normal”
        out.WriteByte(c)
        lastSpaceOutsideString = false
    }

    s := out.String()
    s = strings.TrimSpace(s)

    // Apertar espaços em torno de símbolos onde é seguro
    replacers := []struct{ old, new string }{
        {" ;", ";"}, {"; ", ";"},
        {" :", ":"}, {": ", ":"},
        {" ,", ","}, {", ", ","},
        {" {", "{"}, {"{ ", "{"},
        {" }", "}"}, {"} ", "}"},
        {" (", "("}, {"( ", "("},
        {" )", ")"}, {") ", ")"},
    }
    for _, r := range replacers {
        s = strings.ReplaceAll(s, r.old, r.new)
    }

    return s
}
