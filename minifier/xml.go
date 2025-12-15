// Author: João Pinto
// Date: 2025-12-15
// Purpose: MinifyXML faz uma minificação conservadora de XML
// License: MIT

package minifier

import "strings"

// MinifyXML faz uma minificação conservadora de XML:
//
// - remove comentários <!-- ... --> se opts.RemoveComments = true
// - colapsa espaços entre atributos se opts.CollapseAttrWhitespace = true
//   (ex: <tag  a="1"   b="2"> → <tag a="1" b="2">)
// - remove nós de texto que sejam *apenas* whitespace (indentação)
//   se opts.CollapseTagWhitespace = true
// - preserva sempre o conteúdo de CDATA, processing instructions e <!DOCTYPE ...>
//
// Não mexe no texto "real" (nós de texto com caracteres não whitespace).
func MinifyXML(input string, opts *XMLOptions) string {
    if opts == nil {
        opts = DefaultXMLOptions()
    }

    var out strings.Builder
    var textBuf strings.Builder

    b := []byte(input)
    n := len(b)

    inTag := false       // dentro de <...>
    inAttr := false      // dentro de valor de atributo "..."
    inComment := false   // dentro de <!-- ... -->
    inCDATA := false     // dentro de <![CDATA[ ... ]]>
    inPI := false        // dentro de <? ... ?>
    inDecl := false      // dentro de <!DOCTYPE ...> ou outras declarações <! ... >
    attrQuote := byte(0)

    // helpers para escrever e manter último byte
    var lastOut byte
    writeByte := func(c byte) {
        out.WriteByte(c)
        lastOut = c
    }
    writeString := func(s string) {
        out.WriteString(s)
        if len(s) > 0 {
            lastOut = s[len(s)-1]
        }
    }

    // flush de texto fora de tags / comentários / CDATA
		flushText := func() {
				if textBuf.Len() == 0 {
						return
				}
				s := textBuf.String()
				if isAllXMLWhitespace(s) {
						// texto é só indentação
						if opts.CollapseTagWhitespace {
								// deitamos fora
						} else {
								writeString(s)
						}
				} else {
						// texto "real"
						if opts.CollapseTagWhitespace {
								// remove apenas whitespace no início/fim,
								// preservando espaços internos (ex: "Texto com  espaços")
								trimmed := strings.Trim(s, " \t\r\n")
								writeString(trimmed)
						} else {
								// modo conservador: não tocar em nada
								writeString(s)
						}
				}
				textBuf.Reset()
		}

    for i := 0; i < n; {
        c := b[i]

        // 1) Comentários <!-- ... -->
        if inComment {
            if i+2 < n && b[i] == '-' && b[i+1] == '-' && b[i+2] == '>' {
                if !opts.RemoveComments {
                    writeString("-->")
                }
                inComment = false
                i += 3
                continue
            }
            if !opts.RemoveComments {
                writeByte(c)
            }
            i++
            continue
        }

        // 2) CDATA <![CDATA[ ... ]]>
        if inCDATA {
            if i+2 < n && b[i] == ']' && b[i+1] == ']' && b[i+2] == '>' {
                writeString("]]>")
                inCDATA = false
                i += 3
                continue
            }
            writeByte(c)
            i++
            continue
        }

        // 3) Processing instruction <? ... ?>
        if inPI {
            if i+1 < n && b[i] == '?' && b[i+1] == '>' {
                writeString("?>")
                inPI = false
                i += 2
                continue
            }
            writeByte(c)
            i++
            continue
        }

        // 4) Declaração genérica <! ... > (ex: <!DOCTYPE ...>)
        if inDecl {
            writeByte(c)
            if c == '>' {
                inDecl = false
            }
            i++
            continue
        }

        // 5) Dentro de <tag ...>
        if inTag {
            if inAttr {
                // dentro de valor de atributo: não mexer
                writeByte(c)
                if c == attrQuote {
                    inAttr = false
                }
                i++
                continue
            }

            switch c {
            case '"', '\'':
                inAttr = true
                attrQuote = c
                writeByte(c)
                i++
            case '>':
                writeByte(c)
                inTag = false
                i++
            case ' ', '\t', '\n', '\r':
                if opts.CollapseAttrWhitespace {
                    // colapsar whitespace entre nome/atributos em um único espaço
                    if lastOut != ' ' && lastOut != '<' {
                        writeByte(' ')
                    }
                } else {
                    writeByte(c)
                }
                i++
            default:
                writeByte(c)
                i++
            }
            continue
        }

        // 6) Fora de tags (texto ou início de markup)
        if c == '<' {
            // texto acumulado até aqui
            flushText()

            // Verificar que tipo de markup é
            if i+1 < n && b[i+1] == '?' {
                // processing instruction: <? ... ?>
                inPI = true
                writeString("<?")
                i += 2
                continue
            }

            if i+3 < n && b[i+1] == '!' && b[i+2] == '-' && b[i+3] == '-' {
                // comentário <!-- ... -->
                inComment = true
                if !opts.RemoveComments {
                    writeString("<!--")
                }
                i += 4
                continue
            }

            if i+8 < n && b[i+1] == '!' && b[i+2] == '[' &&
                b[i+3] == 'C' && b[i+4] == 'D' && b[i+5] == 'A' &&
                b[i+6] == 'T' && b[i+7] == 'A' && b[i+8] == '[' {
                // <![CDATA[
                inCDATA = true
                writeString("<![CDATA[")
                i += 9
                continue
            }

            if i+1 < n && b[i+1] == '!' {
                // outra declaração <! ...> (ex: <!DOCTYPE ...>)
                inDecl = true
                writeString("<!")
                i += 2
                continue
            }

            // Caso normal: tag de elemento (abertura/fecho/empty)
            inTag = true
            writeByte('<')
            i++
            continue
        }

        // Texto normal (fora de qualquer markup especial)
        textBuf.WriteByte(c)
        i++
    }

    // flush de texto final
    flushText()

    return out.String()
}

// isAllXMLWhitespace devolve true se a string for apenas espaço/tab/newline.
func isAllXMLWhitespace(s string) bool {
    for i := 0; i < len(s); i++ {
        switch s[i] {
        case ' ', '\t', '\n', '\r':
            continue
        default:
            return false
        }
    }
    return true
}
