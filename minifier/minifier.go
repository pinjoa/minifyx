// Author: João Pinto
// Date: 2025-12-15
// Purpose: módulo principal da biblioteca, definição das configurações por omissão, funções de apoio ao tratamento do conteúdo
// License: MIT

package minifier

import (
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
)

type XMLOptions struct {
    RemoveComments         bool // <!-- ... -->
    CollapseAttrWhitespace bool // múltiplos espaços entre atributos → um
    CollapseTagWhitespace  bool // remover whitespace entre tags, se for só whitespace
    PreserveCDATA          bool // por defeito true
}

func DefaultXMLOptions() *XMLOptions {
    return &XMLOptions{
        RemoveComments:         true,
        CollapseAttrWhitespace: true,
        CollapseTagWhitespace:  true,
        PreserveCDATA:          true,
    }
}

type Options struct {
    // --- Comentários HTML ---

    // remover comentários HTML em geral (<!-- ... -->)
    RemoveHTMLComments bool

    // preservar comentários condicionais tipo <!--[if lt IE 9]> ... <![endif]-->
    PreserveConditionalComments bool

    // preservar comentários de licença tipo <!--! ... -->
    PreserveLicenseComments bool

    // --- Blocos especiais de texto/código ---

    // tratar <pre>/<code> como blocos especiais (não passar pelo minificador normal de HTML)
    PreservePre bool

    // em <pre>: remover apenas whitespace à direita, mantendo indentação à esquerda
    TrimPreRight bool

    // em <code>: minificar o conteúdo numa única linha (colapsar whitespace interno)
    MinifyCodeBlocks bool

    // em <textarea>: aplicar trim ao conteúdo (remover whitespace no início/fim)
    MinifyTextarea bool

    // --- Templates HTML / scripts de template ---

    // minificar HTML dentro de <template>...</template>
    MinifyHTMLTemplates bool

    // minificar HTML dentro de <script type="text/html|text/x-handlebars-template|text/x-template">
    MinifyScriptTemplates bool

    // --- CSS / JS / JSON embebidos ---

    // minificar CSS dentro de <style>...</style> usando MinifyCSS
    MinifyInlineCSS bool

    // minificar JS dentro de <script>...</script> (scripts "normais") usando MinifyJS
    MinifyInlineJS bool

    // minificar JSON dentro de <script type="application/ld+json|application/json">
    // usando MinifyJSON
    MinifyJSONScripts bool

    // minificar JSON em atributos data-json="..." / data-json='...' usando MinifyJSON
    MinifyDataJSON bool

    // --- Whitespace & espaçamentos no HTML "de fora" ---

    // aplicar o minificador de whitespace HTML-aware (minifyHTMLWhitespace)
    // em texto e tags fora dos blocos protegidos
    CollapseHTMLWhitespace bool

    // preservar um espaço entre tags inline (ex: </span> <span>) para não colar texto
    PreserveInlineTagSpaces bool

    // remover espaços entre tags de "bloco" (ex: </div> <script> -> </div><script>)
    TightenBlockTagGaps bool
}

func DefaultOptions() *Options {
    return &Options{
        // Comentários
        RemoveHTMLComments:        true,
        PreserveConditionalComments: false,
        PreserveLicenseComments:   false,

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
    }
}

// Minify por tipo
func Minify(input string, t Type, hopts *Options, xopts *XMLOptions) string {
    switch t { 
    case HTML:
        if hopts == nil { hopts = DefaultOptions() }
        return MinifyHTML(input, hopts)
    case CSS:
        return MinifyCSS(input)
    case JS:
        return MinifyJS(input)
    case JSON:
        return MinifyJSON(input)
    case XML:
        if xopts == nil { xopts = DefaultXMLOptions() }
        return MinifyXML(input, xopts)
    default:
        return input
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
        return HTML
    }
}

// MinifyFile lê, deteta tipo e minifica
func MinifyFile(path string, hopts *Options, xopts *XMLOptions) (string, error) {
    b, err := os.ReadFile(path)
    if err != nil { return "", err }
    t := DetectType(path)
    return Minify(string(b), t, hopts, xopts), nil
}

// MinifyReader lê de io.Reader e minifica conforme tipo
func MinifyReader(r io.Reader, t Type, hopts *Options, xopts *XMLOptions) (string, error) {
    b, err := io.ReadAll(r)
    if err != nil { return "", err }
    return Minify(string(b), t, hopts, xopts), nil
}

// MinifyToWriter lê de reader, minifica e escreve em writer
func MinifyToWriter(r io.Reader, w io.Writer, t Type, hopts *Options, xopts *XMLOptions) error {
    out, err := MinifyReader(r, t, hopts, xopts)
    if err != nil { return err }
    _, err = io.WriteString(w, out)
    return err
}
