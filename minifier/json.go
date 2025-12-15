// Author: João Pinto
// Date: 2025-12-15
// Purpose: MinifyJSON remove todo o whitespace fora de strings. Assume JSON válido e não altera nada dentro de strings.
// License: MIT

package minifier

func MinifyJSON(input string) string {
    var out []byte
    inString := false
    escaped := false

    b := []byte(input)

    for i := 0; i < len(b); i++ {
        c := b[i]

        if inString {
            out = append(out, c)

            if escaped {
                // este char está escapado, voltamos ao normal
                escaped = false
                continue
            }

            if c == '\\' {
                // próximo char será escapado
                escaped = true
                continue
            }

            if c == '"' {
                // fim da string
                inString = false
            }

            continue
        }

        // fora de string
        switch c {
        case ' ', '\n', '\r', '\t':
            // ignorar todo o whitespace fora de strings
            continue
        case '"':
            inString = true
            escaped = false
            out = append(out, c)
        default:
            out = append(out, c)
        }
    }

    return string(out)
}
