// Author: João Pinto
// Date: 2025-12-15
// Purpose: módulo principal da biblioteca, definição das configurações por omissão, funções de apoio ao tratamento do conteúdo
// License: MIT

package minifier

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Type int

const (
    HTML Type = iota
    CSS
    JS
    JSON
    XML
    ERROR
)

type Options struct {
    // --- Comentários HTML ---
    // remover comentários HTML em geral (<!-- ... -->)
    RemoveHTMLComments          bool
    // preservar comentários condicionais tipo <!--[if lt IE 9]> ... <![endif]-->
    PreserveConditionalComments bool
    // preservar comentários de licença tipo <!--! ... -->
    PreserveLicenseComments     bool

    // --- Blocos especiais de texto/código ---
    // tratar <pre>/<code> como blocos especiais (não passar pelo minificador normal de HTML)
    PreservePre      bool
    // em <pre>: remover apenas whitespace à direita, mantendo indentação à esquerda
    TrimPreRight     bool
    // em <code>: minificar o conteúdo numa única linha (colapsar whitespace interno)
    MinifyCodeBlocks bool
    // em <textarea>: aplicar trim ao conteúdo (remover whitespace no início/fim)
    MinifyTextarea   bool

    // --- Templates HTML / scripts de template ---
    // minificar HTML dentro de <template>...</template>
    MinifyHTMLTemplates   bool
    // minificar HTML dentro de <script type="text/html|text/x-handlebars-template|text/x-template">
    MinifyScriptTemplates bool

    // --- CSS / JS / JSON embebidos ---
    // minificar CSS dentro de <style>...</style> usando MinifyCSS
    MinifyInlineCSS   bool
    // minificar JS dentro de <script>...</script> (scripts "normais") usando MinifyJS
    MinifyInlineJS    bool
    // minificar JSON dentro de <script type="application/ld+json|application/json">
    // usando MinifyJSON
    MinifyJSONScripts bool
    // minificar JSON em atributos data-json="..." / data-json='...' usando MinifyJSON
    MinifyDataJSON    bool

    // --- Whitespace & espaçamentos no HTML "de fora" ---
    // aplicar o minificador de whitespace HTML-aware (minifyHTMLWhitespace)
    // em texto e tags fora dos blocos protegidos
    CollapseHTMLWhitespace  bool
    // preservar um espaço entre tags inline (ex: </span> <span>) para não colar texto
    PreserveInlineTagSpaces bool
    // remover espaços entre tags de "bloco" (ex: </div> <script> -> </div><script>)
    TightenBlockTagGaps     bool

    // --- relacionado apenas com XML ---
    XMLRemoveComments         bool // <!-- ... -->
    XMLCollapseAttrWhitespace bool // múltiplos espaços entre atributos → um
    XMLCollapseTagWhitespace  bool // remover whitespace entre tags, se for só whitespace
    XMLPreserveCDATA          bool // por defeito true

}

func DefaultOptions() *Options {
    return &Options{
        // Comentários
        RemoveHTMLComments:          true,
        PreserveConditionalComments: false,
        PreserveLicenseComments:     false,

        // Blocos especiais
        PreservePre:       true,
        TrimPreRight:      true,
        MinifyCodeBlocks:  false,
        MinifyTextarea:    true,

        // Templates
        MinifyHTMLTemplates:   true,
        MinifyScriptTemplates: true,

        // CSS / JS / JSON embebidos
        MinifyInlineCSS:   true,
        MinifyInlineJS:    true,
        MinifyJSONScripts: true,
        MinifyDataJSON:    true,

        // Whitespace HTML
        CollapseHTMLWhitespace:  true,
        PreserveInlineTagSpaces: true,
        TightenBlockTagGaps:     true,

        // XML apenas
        XMLRemoveComments:         true,
        XMLCollapseAttrWhitespace: true,
        XMLCollapseTagWhitespace:  true,
        XMLPreserveCDATA:          true,
    }
}

// Minify por tipo
func Minify(input string, t Type, opts *Options) (string, error) {
    switch t { 
    case HTML:
        if opts == nil { opts = DefaultOptions() }
        return MinifyHTML(input, opts), nil
    case CSS:
        return MinifyCSS(input), nil
    case JS:
        return MinifyJS(input), nil
    case JSON:
        return MinifyJSON(input), nil
    case XML:
        if opts == nil { opts = DefaultOptions() }
        return MinifyXML(input, opts), nil
    default:
        return "", errors.New("tipo não suportado")
    }
}

// Detecta tipo a partir da extensão
func DetectType(path string) Type {
    switch strings.ToLower(filepath.Ext(path)) {
    case ".html", ".htm":
        return HTML
    case ".css":
        return CSS
    case ".js":
        return JS
    case ".json":
        return JSON
    case ".xml":
        return XML
    default:
        return ERROR
    }
}

// MinifyFile lê, deteta tipo e minifica
func MinifyFile(path string, opts *Options) (string, error) {
    b, err := os.ReadFile(path)
    if err != nil { return "", err }
    t := DetectType(path)
    if t == ERROR { return "", errors.New("tipo não suportado") }
    return Minify(string(b), t, opts)
}

// MinifyReader lê de io.Reader e minifica conforme tipo
func MinifyReader(r io.Reader, t Type, opts *Options) (string, error) {
    if t == ERROR { return "", errors.New("tipo não suportado") }
    b, err := io.ReadAll(r)
    if err != nil { return "", err }
    return Minify(string(b), t, opts)
}

// MinifyToWriter lê de reader, minifica e escreve em writer
func MinifyToWriter(r io.Reader, w io.Writer, t Type, opts *Options) error {
    if t == ERROR { return errors.New("não suportado") }
    out, err := MinifyReader(r, t, opts)
    if err != nil { return err }
    _, err = io.WriteString(w, out)
    return err
}
