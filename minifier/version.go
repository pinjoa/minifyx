package minifier

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

// Version é a versão atual da biblioteca; é substituída em build via ldflags.
var (
    Version   = "dev"
    Commit    = ""
    BuildDate = ""
)

// ResolvedVersion devolve a versão efetiva da biblioteca.
func ResolvedVersion() string {
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

// VersionInfo devolve string formatada para o CLI.
func VersionInfo() string {
    return fmt.Sprintf(
        "minifyx version %s\ncommit:    %s\nbuilt at:  %s\ngo:        %s\nplatform:  %s/%s",
        Version,
        Commit,
        BuildDate,
        runtime.Version(),
        runtime.GOOS,
        runtime.GOARCH,
    )
}
