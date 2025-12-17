// Author: João Pinto
// Date: 2025-12-15
// Purpose: módulo principal da aplicação CLI
// License: MIT

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
    "runtime/debug"

	"github.com/pinjoa/minifyx/minifier"
)

func resolvedVersion() string {
    // 1) se o release.yml injectou a versão, usa-a
    if minifier.Version != "" && minifier.Version != "dev" {
        return minifier.Version
    }
    // 2) caso contrário, tenta ler a versão do módulo (go install @vX.Y.Z)
    if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" {
        return bi.Main.Version
    }
    // 3) fallback
    return "dev"
}

func main() {
    var (
        showVersion bool
        outPath     string
        useStdout   bool
        useStdin    bool
        forceType   string
        parallel    int

        // opções HTML
        preservePreCode    bool // controla tratamento especial de <pre>/<code>
        removeHTMLComments bool // remover comentários <!-- ... -->

        disableHTMLWhitespace bool // não colapsar espaços em HTML "de fora"
        disableHTMLTemplates  bool // não minificar <template> e scripts de template
        disableHTMLJSON       bool // não minificar JSON em <script type="application/*json"> e data-json
        
        // opções XML (novas)
        removeXMLComments bool
        noXMLWhitespace   bool
    )

    flag.BoolVar(&showVersion, "version", false, "mostrar versão")
    flag.StringVar(&outPath, "o", "", "Saída (ficheiro ou diretoria)")
    flag.BoolVar(&useStdout, "stdout", false, "Escrever para stdout")
    flag.BoolVar(&useStdin, "stdin", false, "Ler de stdin")
    flag.StringVar(&forceType, "type", "", "Forçar tipo: html|css|js|json|xml")
    flag.IntVar(&parallel, "parallel", runtime.NumCPU(), "Número de goroutines em paralelo")

    flag.BoolVar(&preservePreCode, "preserve-precode", true, "Preservar conteúdo especial em <pre>/<code> (e tratar <code> como bloco)")
    flag.BoolVar(&removeHTMLComments, "remove-html-comments", true, "Remover comentários HTML")

    flag.BoolVar(&disableHTMLWhitespace, "no-html-whitespace", false, "Não colapsar espaços em HTML (texto fora de blocos especiais)")
    flag.BoolVar(&disableHTMLTemplates, "no-html-templates", false, "Não minificar HTML dentro de <template> e scripts de template")
    flag.BoolVar(&disableHTMLJSON, "no-html-json", false, "Não minificar JSON em <script type=\"application/*json\"> e atributos data-json")

    flag.BoolVar(&removeXMLComments, "remove-xml-comments", true, "Remover comentários XML")
    flag.BoolVar(&noXMLWhitespace, "no-xml-whitespace", false, "Não colapsar espaços/indentação em XML")

    flag.Parse()

    if showVersion {
        fmt.Printf("minifyx CLI: %s\n", resolvedVersion())
        fmt.Printf("minifyx lib: %s\n", minifier.Version)
        return
    }

    opts := minifier.DefaultOptions()
    opts.XMLRemoveComments = removeXMLComments
    if noXMLWhitespace {
        opts.XMLCollapseTagWhitespace =  false
        opts.XMLCollapseAttrWhitespace = false
    }

    // Comentários HTML
    opts.RemoveHTMLComments = removeHTMLComments

    // Tratamento de <pre>/<code>/<textarea>
    if preservePreCode {
        // comportamento “rico” que já validaste:
        opts.PreservePre =      true  // <pre> protegido
        opts.TrimPreRight =     true  // corta lixo no fim do <pre>
        opts.MinifyCodeBlocks = true  // <code> numa linha, whitespace colapsado
        // MinifyTextarea fica com default true
    } else {
        // modo mais “cru”: não tratar <pre> de forma especial
        opts.PreservePre =      false
        opts.TrimPreRight =     false
        // ainda tratamos <code> como bloco, mas sem apertar o conteúdo
        opts.MinifyCodeBlocks = false
    }

    // Templates HTML & scripts de template
    if disableHTMLTemplates {
        opts.MinifyHTMLTemplates =   false
        opts.MinifyScriptTemplates = false
    }

    // JSON (scripts + data-json)
    if disableHTMLJSON {
        opts.MinifyJSONScripts = false
        opts.MinifyDataJSON =    false
    }

    // Whitespace HTML “de fora”
    if disableHTMLWhitespace {
        opts.CollapseHTMLWhitespace = false
        opts.TightenBlockTagGaps =    false
    }

    if useStdin {
        reader := bufio.NewReader(os.Stdin)
        var b strings.Builder
        for {
            chunk, err := reader.ReadString('\n')
            b.WriteString(chunk)
            if err == io.EOF {
                break
            }
            if err != nil {
                fmt.Fprintln(os.Stderr, "Erro ao ler stdin:", err)
                os.Exit(1)
            }
        }
        input := b.String()
        var t minifier.Type
        switch strings.ToLower(forceType) {
        case "html":
            t = minifier.HTML
        case "css":
            t = minifier.CSS
        case "js":
            t = minifier.JS
        case "json":
            t = minifier.JSON
        case "xml":
            t = minifier.XML
        default:
            fmt.Fprintln(os.Stderr, "É necessário -type quando usa -stdin (html|css|js|json|xml)")
            os.Exit(2)
        }
        out, err := minifier.Minify(input, t, opts)
        if err != nil {
            fmt.Fprintln(os.Stderr, err)
            os.Exit(2)
        }
        if useStdout || outPath == "" {
            fmt.Print(out)
        } else {
            if err := os.WriteFile(outPath, []byte(out), 0644); err != nil {
                fmt.Fprintln(os.Stderr, "Erro a escrever saída:", err)
                os.Exit(1)
            }
        }
        return
    }

    args := flag.Args()
    if len(args) == 0 {
        fmt.Println("Uso: minifyx [opções] <ficheiros...>\n\nou: minifyx -help\n\nEx.: minifyx -parallel 4 index.html style.css app.js")
        os.Exit(0)
    }

    type job struct { path string }
    type result struct { path string; out string; err error }

    jobs := make(chan job)
    results := make(chan result)

    for i := 0; i < parallel; i++ {
        go func() {
            for j := range jobs {
                out, err := minifier.MinifyFile(j.path, opts)
                results <- result{path: j.path, out: out, err: err}
            }
        }()
    }

    go func() {
        for _, p := range args {
            jobs <- job{path: p}
        }
        close(jobs)
    }()

    pending := len(args)
    for pending > 0 {
        r := <-results
        pending--
        if r.err != nil {
            fmt.Fprintln(os.Stderr, "Erro:", r.path, r.err)
            continue
        }
        if useStdout {
            fmt.Println(r.out)
            continue
        }
        dest := outPath
        if dest == "" {
            ext := filepath.Ext(r.path)
            base := strings.TrimSuffix(r.path, ext)
            dest = base + ".min" + ext
        } else {
            info, _ := os.Stat(dest)
            if info != nil && info.IsDir() {
                ext := filepath.Ext(r.path)
                name := filepath.Base(strings.TrimSuffix(r.path, ext))
                if forceType != "" {
                    ext = "." + forceType
                }
                dest = filepath.Join(dest, name+".min"+ext)
            }
        }
        if err := os.WriteFile(dest, []byte(r.out), 0644); err != nil {
            fmt.Fprintln(os.Stderr, "Erro a escrever:", dest, err)
        } else {
            fmt.Println("Minificado:", dest)
        }
    }
}
