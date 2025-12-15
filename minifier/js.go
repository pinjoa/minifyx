// Author: João Pinto
// Date: 2025-12-15
// Purpose: MinifyJS remove comentários respeitando strings, templates e regex, 
//          tenta manter tudo numa linha (exceto newlines lógicos dentro de templates, 
//          que são convertidos para '\n') e depois aperta espaços desnecessários fora de literais.
// License: MIT

package minifier

import (
	"strings"
)

func MinifyJS(input string) string {
    const (
        modeNormal = iota
        modeLineComment
        modeBlockComment
        modeString
        modeTemplate
        modeRegex
    )

    var out strings.Builder
    b := []byte(input)

    mode := modeNormal
    escaped := false
    stringQuote := byte(0)
    inCharClass := false

    var wordBuf []byte
    lastWord := ""
    prevNonSpace := byte(0)
    asiKeywordPending := false
    lastOut := byte(0)

    // 1º passe: remover comentários, newlines “soltos”, etc.
    for i := 0; i < len(b); i++ {
        c := b[i]
        var next byte
        if i+1 < len(b) {
            next = b[i+1]
        } else {
            next = 0
        }

        switch mode {
        case modeLineComment:
            // ignora até newline; se for depois de return/throw/break/continue,
            // insere ';' (ASI explícito)
            if c == '\n' || c == '\r' {
                mode = modeNormal
                if asiKeywordPending {
                    if lastOut != ';' {
                        out.WriteByte(';')
                        lastOut = ';'
                        prevNonSpace = ';'
                    }
                    asiKeywordPending = false
                }
            }
            continue

        case modeBlockComment:
            if c == '*' && next == '/' {
                mode = modeNormal
                i++
            }
            continue

        case modeString:
            out.WriteByte(c)
            lastOut = c
            if escaped {
                escaped = false
                continue
            }
            if c == '\\' {
                escaped = true
                continue
            }
            if c == stringQuote {
                mode = modeNormal
            }
            continue

        case modeTemplate:
            // Template literal: copiamos tudo tal como está,
            // exceto que newlines físicos viram sequências "\n"
            if escaped {
                out.WriteByte(c)
                lastOut = c
                escaped = false
                continue
            }

            if c == '\\' {
                out.WriteByte(c)
                lastOut = c
                escaped = true
                continue
            }

            if c == '\n' || c == '\r' {
                // converte newline em '\n'
                out.WriteByte('\\')
                out.WriteByte('n')
                lastOut = 'n'
                // tratar CRLF como um único newline lógico
                if c == '\r' && next == '\n' {
                    i++
                }
                continue
            }

            out.WriteByte(c)
            lastOut = c
            if c == '`' {
                mode = modeNormal
            }
            continue

        case modeRegex:
            out.WriteByte(c)
            lastOut = c
            if escaped {
                escaped = false
                continue
            }
            if c == '\\' {
                escaped = true
                continue
            }
            if c == '[' {
                inCharClass = true
                continue
            }
            if c == ']' {
                inCharClass = false
                continue
            }
            if c == '/' && !inCharClass {
                mode = modeNormal
                prevNonSpace = '/'
            }
            continue
        }

        // modeNormal

        // palavra/identificador (para ASI e regex)
        if isIdentChar(c) {
            out.WriteByte(c)
            lastOut = c
            wordBuf = append(wordBuf, c)
            prevNonSpace = c
            continue
        } else {
            if len(wordBuf) > 0 {
                lastWord = string(wordBuf)
                if isASIKeyword(lastWord) {
                    asiKeywordPending = true
                } else {
                    asiKeywordPending = false
                }
                wordBuf = wordBuf[:0]
            }
        }

        // Comentários ou regex/divisão
        if c == '/' {
            if next == '/' {
                mode = modeLineComment
                // não limpamos asiKeywordPending, por causa de:
                // return // coment
                // 1;
                i++
                continue
            }
            if next == '*' {
                mode = modeBlockComment
                i++
                continue
            }

            // Possível início de regex
            if isRegexStart(prevNonSpace, lastWord) {
                mode = modeRegex
                escaped = false
                inCharClass = false
                out.WriteByte(c)
                lastOut = c
                asiKeywordPending = false
                continue
            }

            // senão, é divisão normal
            out.WriteByte(c)
            lastOut = c
            prevNonSpace = c
            asiKeywordPending = false
            continue
        }

        // Strings
        if c == '\'' || c == '"' {
            mode = modeString
            stringQuote = c
            escaped = false
            out.WriteByte(c)
            lastOut = c
            prevNonSpace = c
            asiKeywordPending = false
            continue
        }

        // Template literal
        if c == '`' {
            mode = modeTemplate
            escaped = false
            out.WriteByte(c)
            lastOut = c
            prevNonSpace = c
            asiKeywordPending = false
            continue
        }

        // Espaços / tabs
        if c == ' ' || c == '\t' {
            // colapsar em um único espaço se fizer sentido
            if lastOut != ' ' && lastOut != 0 {
                out.WriteByte(' ')
                lastOut = ' '
            }
            continue
        }

        // Newlines fora de strings/templates/regex
        if c == '\n' || c == '\r' {
            if asiKeywordPending {
                // depois de return/throw/break/continue -> mete ';'
                if lastOut != ';' {
                    out.WriteByte(';')
                    lastOut = ';'
                    prevNonSpace = ';'
                }
                asiKeywordPending = false
            }
            // senão, simplesmente não escreve newline
            continue
        }

        // Qualquer outro char
        out.WriteByte(c)
        lastOut = c
        if !isWhitespace(c) {
            prevNonSpace = c
            asiKeywordPending = false
        }
    }

    // 2º passe: apertar espaços fora de strings/templates/regex
    result := out.String()
    result = tightenSpacesJS(result)
    return result
}

// Segundo passe: remove espaços desnecessários fora de strings/templates/regex.
func tightenSpacesJS(code string) string {
    const (
        modeNormal = iota
        modeString
        modeTemplate
        modeRegex
    )

    var out strings.Builder
    b := []byte(code)

    mode := modeNormal
    escaped := false
    stringQuote := byte(0)
    inCharClass := false

    var wordBuf []byte
    lastWord := ""
    prevNonSpace := byte(0)
    lastOut := byte(0)

    for i := 0; i < len(b); i++ {
        c := b[i]

        switch mode {
        case modeString:
            out.WriteByte(c)
            lastOut = c
            if escaped {
                escaped = false
                continue
            }
            if c == '\\' {
                escaped = true
                continue
            }
            if c == stringQuote {
                mode = modeNormal
            }
            continue

        case modeTemplate:
            // copiar exatamente (inclui '\n' como sequência \n, não newline real)
            out.WriteByte(c)
            lastOut = c
            if escaped {
                escaped = false
                continue
            }
            if c == '\\' {
                escaped = true
                continue
            }
            if c == '`' {
                mode = modeNormal
            }
            continue

        case modeRegex:
            out.WriteByte(c)
            lastOut = c
            if escaped {
                escaped = false
                continue
            }
            if c == '\\' {
                escaped = true
                continue
            }
            if c == '[' {
                inCharClass = true
                continue
            }
            if c == ']' {
                inCharClass = false
                continue
            }
            if c == '/' && !inCharClass {
                mode = modeNormal
                prevNonSpace = '/'
            }
            continue
        }

        // modeNormal

        // palavra/identificador
        if isIdentChar(c) {
            out.WriteByte(c)
            lastOut = c
            wordBuf = append(wordBuf, c)
            prevNonSpace = c
            continue
        } else {
            if len(wordBuf) > 0 {
                lastWord = string(wordBuf)
                wordBuf = wordBuf[:0]
            }
        }

        // regex/divisão
        if c == '/' {
            if isRegexStart(prevNonSpace, lastWord) {
                mode = modeRegex
                escaped = false
                inCharClass = false
                out.WriteByte(c)
                lastOut = c
                continue
            }
            out.WriteByte(c)
            lastOut = c
            prevNonSpace = c
            continue
        }

        // strings
        if c == '"' || c == '\'' {
            mode = modeString
            stringQuote = c
            escaped = false
            out.WriteByte(c)
            lastOut = c
            prevNonSpace = c
            continue
        }

        // template
        if c == '`' {
            mode = modeTemplate
            escaped = false
            out.WriteByte(c)
            lastOut = c
            prevNonSpace = c
            continue
        }

        // whitespace (espaço, tabs, eventuais newlines residuais)
        if isWhitespace(c) {
            // procurar próximo char não-whitespace
            j := i + 1
            var nextNon byte
            for j < len(b) {
                if !isWhitespace(b[j]) {
                    nextNon = b[j]
                    break
                }
                j++
            }
            if nextNon == 0 {
                // só lixo no fim -> ignora
                continue
            }

            // decidir se este espaço é dispensável
            if shouldSkipSpace(prevNonSpace, nextNon) {
                continue
            }

            if lastOut != ' ' && lastOut != 0 {
                out.WriteByte(' ')
                lastOut = ' '
            }
            continue
        }

        // resto
        out.WriteByte(c)
        lastOut = c
        if !isWhitespace(c) {
            prevNonSpace = c
        }
    }

    return out.String()
}

// Espaços que podemos tirar à vontade à volta destes símbolos
func isNoSpacePunct(c byte) bool {
    switch c {
    case '(', ')', '{', '}', '[', ']', ';', ',', ':',
        '*', '/', '%',
        '&', '|', '^', '!', '~',
        '?',
        '=', '<', '>':
        return true
    default:
        return false
    }
}

func shouldSkipSpace(prev, next byte) bool {
    if prev == 0 {
        return true
    }
    if isNoSpacePunct(prev) {
        return true
    }
    if isNoSpacePunct(next) {
        return true
    }
    return false
}

func isIdentChar(c byte) bool {
    return (c >= 'a' && c <= 'z') ||
        (c >= 'A' && c <= 'Z') ||
        (c >= '0' && c <= '9') ||
        c == '_' || c == '$'
}

func isWhitespace(c byte) bool {
    return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// Palavras que podem acionar ASI se vier um newline a seguir.
func isASIKeyword(w string) bool {
    switch w {
    case "return", "throw", "break", "continue":
        return true
    default:
        return false
    }
}

// Heurística para detetar início de regex literal.
func isRegexStart(prevNonSpace byte, lastWord string) bool {
    if prevNonSpace == 0 {
        return true
    }

    const syms = "=({[,!?:;&|^~<>+-*%"

    if strings.ContainsRune(syms, rune(prevNonSpace)) {
        return true
    }

    switch lastWord {
    case "", "return", "case", "throw", "else", "do",
        "typeof", "instanceof", "delete", "void", "in", "of",
        "yield", "await", "new", "while", "for", "if", "catch", "switch":
        return true
    }

    return false
}
