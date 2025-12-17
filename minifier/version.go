package minifier

import "runtime/debug"

// Version é a versão atual da biblioteca; é substituída em build via ldflags.
var Version = "dev"

// ResolvedVersion devolve a versão efetiva da biblioteca.
func ResolvedVersion() string {
    if Version != "" && Version != "dev" {
        return Version
    }
    if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" {
        return bi.Main.Version
    }
    return "dev"
}
