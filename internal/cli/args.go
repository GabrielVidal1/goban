package cli

import (
	"flag"
	"os"
	"strings"
)

// GlobalFlags holds flags that apply across all commands.
type GlobalFlags struct {
	Dir  string
	Host string
	Port string
}

// ParseTopLevel separates global flags from the remaining args (command + its
// own flags/args). Global flags may appear anywhere before "--".
func ParseTopLevel(args []string) (GlobalFlags, []string) {
	fs := flag.NewFlagSet("kanban-ui", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var g GlobalFlags
	fs.StringVar(&g.Dir, "dir", os.Getenv("KANBAN_DIR"), "kanban directory (overrides KANBAN_DIR)")
	fs.StringVar(&g.Host, "host", os.Getenv("HOST"), "listen host (overrides HOST)")
	fs.StringVar(&g.Port, "port", os.Getenv("PORT"), "listen port (overrides PORT)")

	// Use the mixed-position parser so flags can precede or follow the command.
	_ = parseFlags(fs, args)
	return g, fs.Args()
}

// addCommonFlags registers --dir (defaulting to the global config) and
// optionally --json on fs.
func addCommonFlags(fs *flag.FlagSet, dir *string, asJSON *bool) {
	defaultDir := ""
	if globalConfig != nil {
		defaultDir = globalConfig.KanbanDir
	}
	fs.StringVar(dir, "dir", defaultDir, "kanban data directory (overrides KANBAN_DIR)")
	if asJSON != nil {
		fs.BoolVar(asJSON, "json", false, "output as JSON")
	}
}

// parseFlags lets flags appear before or after positional arguments.
// Go's flag package stops at the first non-flag token; this reorders args so
// flags always come first, then positionals.
func parseFlags(fs *flag.FlagSet, args []string) error {
	type boolFlag interface{ IsBoolFlag() bool }
	var flags, positionals []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			positionals = append(positionals, args[i+1:]...)
			break
		}
		if len(a) > 1 && a[0] == '-' {
			name := strings.TrimLeft(a, "-")
			if strings.Contains(name, "=") {
				flags = append(flags, a)
				continue
			}
			flags = append(flags, a)
			if f := fs.Lookup(name); f != nil {
				if bv, ok := f.Value.(boolFlag); ok && bv.IsBoolFlag() {
					continue
				}
			}
			if i+1 < len(args) {
				flags = append(flags, args[i+1])
				i++
			}
			continue
		}
		positionals = append(positionals, a)
	}
	return fs.Parse(append(flags, positionals...))
}
