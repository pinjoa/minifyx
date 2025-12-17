// Author: João Pinto
// Date: 2025-12-15
// Purpose: MinifyHTML minifica HTML preservando blocos sensíveis,
//          minificando JS, CSS e JSON, e usando um minificador de whitespace
//          com noção de contexto.
// License: MIT

package minifier

import (
	"regexp"
	"strconv"
	"strings"
)

func MinifyHTML(input string, opts *Options) string {
    if opts == nil {
        opts = DefaultOptions()
    }

    html := input

    // 1) Remover comentários HTML (se estiver ativo)
    if opts.RemoveHTMLComments {
        reComments := regexp.MustCompile(`<!--[\s\S]*?-->`)
        html = reComments.ReplaceAllString(html, "")
    }

    // 2) Preparar placeholders para blocos que não devem ser tocados pelo
    //    minificador de whitespace (depois de processados).
    placeholders := []string{}
    blocks := []string{}
    idx := 0

    addPlaceholder := func(block string) string {
        ph := "###HOLD_BLOCK_" + strconv.Itoa(idx) + "###"
        placeholders = append(placeholders, ph)
        blocks = append(blocks, block)
        idx++
        return ph
    }

    // 3) <pre>...</pre> → preservar conteúdo, opcionalmente cortar apenas newline final
    if opts.PreservePre {
        rePre := regexp.MustCompile(`(?is)(<pre[^>]*>)(.*?)(</pre>)`)
        html = rePre.ReplaceAllStringFunc(html, func(m string) string {
            parts := rePre.FindStringSubmatch(m)
            if len(parts) != 4 {
                return m
            }
            inner := parts[2]
            if opts.TrimPreRight {
                inner = trimTrailingNewline(inner)
            }
            block := parts[1] + inner + parts[3]
            return addPlaceholder(block)
        })
    }

    // 4) <code>...</code> → opcionalmente colapsar whitespace para uma linha + placeholder
    reCode := regexp.MustCompile(`(?is)(<code[^>]*>)(.*?)(</code>)`)
    html = reCode.ReplaceAllStringFunc(html, func(m string) string {
        parts := reCode.FindStringSubmatch(m)
        if len(parts) != 4 {
            return m
        }
        inner := parts[2]
        if opts.MinifyCodeBlocks {
            inner = minifyPlainTextSingleLine(inner)
        }
        block := parts[1] + inner + parts[3]
        return addPlaceholder(block)
    })

    // 5) <textarea>...</textarea> → tratar de forma conservadora:
    //    só tiramos whitespace de início/fim do conteúdo se MinifyTextarea = true,
    //    e protegemos sempre com placeholder.
    reTextarea := regexp.MustCompile(`(?is)(<textarea[^>]*>)(.*?)(</textarea>)`)
    html = reTextarea.ReplaceAllStringFunc(html, func(m string) string {
        parts := reTextarea.FindStringSubmatch(m)
        if len(parts) != 4 {
            return m
        }
        inner := parts[2]
        if opts.MinifyTextarea {
            inner = strings.TrimSpace(inner)
        }
        block := parts[1] + inner + parts[3]
        return addPlaceholder(block)
    })

    // 6) <template>...</template> → opcionalmente minificar HTML interno + placeholder
    reTemplate := regexp.MustCompile(`(?is)(<template[^>]*>)(.*?)(</template>)`)
    html = reTemplate.ReplaceAllStringFunc(html, func(m string) string {
        parts := reTemplate.FindStringSubmatch(m)
        if len(parts) != 4 {
            return m
        }
        inner := parts[2]
        if opts.MinifyHTMLTemplates {
            inner = minifyHTMLWhitespace(inner)
        }
        block := parts[1] + inner + parts[3]
        return addPlaceholder(block)
    })

    // 7) Scripts de template: type="text/html", "text/x-handlebars-template",
    //    "text/x-template" → opcionalmente minificar HTML interno, depois placeholder
    reScriptTemplate := regexp.MustCompile(
        `(?is)(<script[^>]*type=["']text/(?:html|x-handlebars-template|x-template)["'][^>]*>)(.*?)(</script>)`,
    )
    html = reScriptTemplate.ReplaceAllStringFunc(html, func(m string) string {
        parts := reScriptTemplate.FindStringSubmatch(m)
        if len(parts) != 4 {
            return m
        }
        inner := parts[2]
        if opts.MinifyScriptTemplates {
            inner = minifyHTMLWhitespace(inner)
        }
        block := parts[1] + inner + parts[3]
        return addPlaceholder(block)
    })

    // 8) <style>...</style> → opcionalmente minificar CSS interno + placeholder
    reStyle := regexp.MustCompile(`(?is)(<style[^>]*>)(.*?)(</style>)`)
    html = reStyle.ReplaceAllStringFunc(html, func(m string) string {
        parts := reStyle.FindStringSubmatch(m)
        if len(parts) != 4 {
            return m
        }
        inner := parts[2]
        if opts.MinifyInlineCSS {
            inner = MinifyCSS(inner)
        }
        block := parts[1] + inner + parts[3]
        return addPlaceholder(block)
    })

    // 9) <script>...</script> → JS ou JSON interno + placeholder
    reScript := regexp.MustCompile(`(?is)(<script[^>]*>)(.*?)(</script>)`)
    html = reScript.ReplaceAllStringFunc(html, func(m string) string {
        parts := reScript.FindStringSubmatch(m)
        if len(parts) != 4 {
            return m
        }

        openTag := strings.ToLower(parts[1])
        inner := parts[2]

        // JSON / JSON-LD
        if strings.Contains(openTag, `type="application/ld+json"`) ||
            strings.Contains(openTag, `type='application/ld+json'`) ||
            strings.Contains(openTag, `type="application/json"`) ||
            strings.Contains(openTag, `type='application/json'`) {

            if opts.MinifyJSONScripts {
                inner = MinifyJSON(inner)
            }
        } else {
            // JS normal
            if opts.MinifyInlineJS {
                inner = MinifyJS(inner)
            }
        }

        block := parts[1] + inner + parts[3]
        return addPlaceholder(block)
    })

    // 10) Minificar JSON em atributos data-json="..." / data-json='...'
    //     usando MinifyJSON (se ativo).
    if opts.MinifyDataJSON {
        reDataJSONDouble := regexp.MustCompile(`(?is)(data-json\s*=\s*")([^"]*)(")`)
        html = reDataJSONDouble.ReplaceAllStringFunc(html, func(m string) string {
            sub := reDataJSONDouble.FindStringSubmatch(m)
            if len(sub) != 4 {
                return m
            }
            minValue := MinifyJSON(sub[2])
            return sub[1] + minValue + sub[3]
        })

        reDataJSONSingle := regexp.MustCompile(`(?is)(data-json\s*=\s*')([^']*)(')`)
        html = reDataJSONSingle.ReplaceAllStringFunc(html, func(m string) string {
            sub := reDataJSONSingle.FindStringSubmatch(m)
            if len(sub) != 4 {
                return m
            }
            minValue := MinifyJSON(sub[2])
            return sub[1] + minValue + sub[3]
        })
    }

    // 11) Minificar whitespace "por fora" (tags + texto)
    if opts.CollapseHTMLWhitespace {
        html = minifyHTMLWhitespace(html)
    }

    // 11.1) Apertar espaços à volta dos blocos protegidos
    if opts.TightenBlockTagGaps {
        html = tightenPlaceholderGaps(html)
    }

    // 12) Restaurar todos os placeholders (pre, code, textarea, templates, style, script, etc.)
    for i, ph := range placeholders {
        html = strings.ReplaceAll(html, ph, blocks[i])
    }

    // 13) Passe final: remover espaços entre tags de bloco
    if opts.TightenBlockTagGaps {
        html = tightenTagGaps(html)
    }

    return html
}

// minifyHTMLWhitespace colapsa whitespace de forma mais inteligente:
//
// - Dentro de tags (<...>):
//   - colapsa espaços entre atributos em um único espaço
//   - não mexe em valores de atributos (entre aspas)
//
// - Fora de tags (texto):
//   - remove indentação “vazia” entre tags (ex: </head> <body> → </head><body>)
//   - mantém espaço entre elementos inline (</span> <span> → espaço fica)
//   - colapsa múltiplos espaços internos em um só
func minifyHTMLWhitespace(s string) string {
    var out strings.Builder
    b := []byte(s)

    inTag := false
    inAttr := false
    attrQuote := byte(0)

    var lastOut byte
    lastTagInline := false

    for i := 0; i < len(b); i++ {
        c := b[i]

        if inTag {
            if inAttr {
                // Dentro do valor de atributo: não mexer
                out.WriteByte(c)
                lastOut = c
                if c == attrQuote {
                    inAttr = false
                }
                continue
            }

            // Dentro de <tag ...>
            switch c {
            case '"', '\'':
                inAttr = true
                attrQuote = c
                out.WriteByte(c)
                lastOut = c
            case '>':
                out.WriteByte(c)
                lastOut = c
                inTag = false
            case ' ', '\n', '\r', '\t':
                // espaço entre atributos → colapsar em um
                if lastOut != ' ' && lastOut != '<' {
                    out.WriteByte(' ')
                    lastOut = ' '
                }
            default:
                out.WriteByte(c)
                lastOut = c
            }
        } else {
            // Fora de tags: texto entre <...> e <...>
            switch c {
            case '<':
                // início de uma tag: vamos ver que tag é para saber se é inline
                _, inline := parseTagName(b, i)
                lastTagInline = inline
                inTag = true
                out.WriteByte(c)
                lastOut = c
            case ' ', '\n', '\r', '\t':
                // colapsar sequência de whitespace
                j := i
                for j < len(b) && (b[j] == ' ' || b[j] == '\n' || b[j] == '\r' || b[j] == '\t') {
                    j++
                }

                var nextNon byte
                if j < len(b) {
                    nextNon = b[j]
                }

                if nextNon == '<' {
                    // whitespace entre tags
                    // vamos ver se a próxima tag é inline
                    _, nextInline := parseTagName(b, j)

                    // se as duas tags forem inline, mantemos um espaço visível
                    if lastTagInline && nextInline {
                        if lastOut != ' ' && lastOut != 0 {
                            out.WriteByte(' ')
                            lastOut = ' '
                        }
                    }
                    // caso contrário, removemos completamente este whitespace
                } else if nextNon != 0 {
                    // whitespace entre texto e mais texto → espaço normal
                    if lastOut != ' ' && lastOut != 0 {
                        out.WriteByte(' ')
                        lastOut = ' '
                    }
                }
                // saltar o resto do whitespace
                i = j - 1
            default:
                out.WriteByte(c)
                lastOut = c
            }
        }
    }

    return strings.TrimSpace(out.String())
}

// minifyPlainTextSingleLine colapsa todo o whitespace (espaços, tabs, newlines)
// para um único espaço e remove espaços no início/fim.
func minifyPlainTextSingleLine(s string) string {
    var out []byte
    inSpace := false

    for i := 0; i < len(s); i++ {
        c := s[i]
        if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
            inSpace = true
            continue
        }
        if inSpace && len(out) > 0 {
            out = append(out, ' ')
        }
        inSpace = false
        out = append(out, c)
    }

    return strings.TrimSpace(string(out))
}

// lista simples de tags inline onde o espaço entre elementos conta visualmente
func isInlineTag(name string) bool {
    switch name {
    case "span", "a", "strong", "em", "b", "i", "small", "label",
        "button", "input", "select", "textarea", "abbr", "cite",
        "code", "s", "u", "sub", "sup", "time":
        return true
    default:
        return false
    }
}

// lê o nome da tag a partir do índice do '<'
func parseTagName(b []byte, ltIndex int) (string, bool) {
    j := ltIndex + 1
    if j >= len(b) {
        return "", false
    }

    // ignorar '/', '!' etc.
    switch b[j] {
    case '/':
        j++
    case '!', '?':
        // <!DOCTYPE ...> ou <!-- ... --> → tratamos como não-inline
        return "", false
    }

    start := j
    for j < len(b) {
        c := b[j]
        if (c >= 'a' && c <= 'z') ||
            (c >= 'A' && c <= 'Z') ||
            (c >= '0' && c <= '9') {
            j++
        } else {
            break
        }
    }

    if j == start {
        return "", false
    }

    name := strings.ToLower(string(b[start:j]))
    return name, isInlineTag(name)
}

// remove espaços desnecessários à volta de blocos protegidos (placeholders)
func tightenPlaceholderGaps(s string) string {
    // >   ###HOLD_BLOCK_n###  →  >###HOLD_BLOCK_n###
    reAfterTag := regexp.MustCompile(`>(\s+)(###HOLD_BLOCK_\d+###)`)
    s = reAfterTag.ReplaceAllString(s, `>$2`)

    // ###HOLD_BLOCK_n###   ###HOLD_BLOCK_m###  →  ###HOLD_BLOCK_n######HOLD_BLOCK_m###
    reBetweenBlocks := regexp.MustCompile(`(###HOLD_BLOCK_\d+###)\s+(###HOLD_BLOCK_\d+###)`)
    s = reBetweenBlocks.ReplaceAllString(s, `$1$2`)

    // ###HOLD_BLOCK_n###   <  →  ###HOLD_BLOCK_n###<
    reBeforeTag := regexp.MustCompile(`(###HOLD_BLOCK_\d+###)\s+<`)
    s = reBeforeTag.ReplaceAllString(s, `$1<`)

    return s
}

// tightenTagGaps remove espaços entre tags de "bloco", por ex.:
// </title> <style>  →  </title><style>
// </code> <textarea> → </code><textarea>
func tightenTagGaps(s string) string {
    // lista de tags que tratamos como "bloco" para este efeito
    // (não inclui <span>, <a>, etc., para não mexer em casos inline)
    re := regexp.MustCompile(
        `>(\s+)<(title\b|style\b|pre\b|code\b|textarea\b|template\b|script\b|div\b|p\b|h[1-6]\b|section\b|article\b|header\b|footer\b|main\b|nav\b|ul\b|ol\b|li\b|table\b|thead\b|tbody\b|tfoot\b|tr\b|td\b|th\b|form\b)`,
    )
    // group 2 é só o nome da tag (por ex. "textarea"), por isso reconstruímos como `><textarea`
    return re.ReplaceAllString(s, `><$2`)
}

// remove apenas uma newline final (\n, \r ou \r\n), preservando espaços antes dela
func trimTrailingNewline(s string) string {
    if strings.HasSuffix(s, "\r\n") {
        return s[:len(s)-2]
    }
    if len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
        return s[:len(s)-1]
    }
    return s
}
