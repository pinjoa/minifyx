package minifier

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

// Version é a versão atual da biblioteca; é substituída em build via ldflags.
var (
    Version   = "dev"
    Commit    = ""
    BuildDate = ""
)

// ResolvedVersion devolve a versão efetiva da biblioteca.
func resolvedVersion() string {
    // testar cenários:
    // 1) se o release.yml injectou a versão, usa-a
    if Version != "" && Version != "dev" {
        return Version
    }
    // 2) caso contrário, tenta ler a versão do módulo compilado
    if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" {
        return bi.Main.Version
    }
    // 3) fallback
    return "dev"
}

// resolvedCommit devolve o commit, se houver.
func resolvedCommit() string {
    // Se for preenchido por ldflags, usa.
    if Commit != "" {
        return Commit
    }
    // Caso contrário, tenta ver em BuildInfo (Go 1.18+, com vcs info)
    if bi, ok := debug.ReadBuildInfo(); ok {
        for _, s := range bi.Settings {
            if s.Key == "vcs.revision" && s.Value != "" {
                return s.Value
            }
        }
    }
    return ""
}

// resolvedBuildDate devolve a data de build, se houver.
func resolvedBuildDate() string {
    if BuildDate != "" {
        return BuildDate
    }
    if bi, ok := debug.ReadBuildInfo(); ok {
        for _, s := range bi.Settings {
            if s.Key == "vcs.time" && s.Value != "" {
                return s.Value
            }
        }
    }
    return ""
}

// VersionInfo devolve string formatada para o CLI.
func VersionInfo() string {
    v := resolvedVersion()
    c := resolvedCommit()
    d := resolvedBuildDate()

    lines := []string{
        fmt.Sprintf("minifyx version %s", v),
    }

    if c != "" {
        lines = append(lines, fmt.Sprintf("commit:    %s", c))
    }
    if d != "" {
        lines = append(lines, fmt.Sprintf("built at:  %s", d))
    }

    lines = append(lines,
        fmt.Sprintf("go:        %s", runtime.Version()),
        fmt.Sprintf("platform:  %s/%s", runtime.GOOS, runtime.GOARCH),
    )

    return strings.Join(lines, "\n")
}
