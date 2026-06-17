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

// ParseTopLevel separates global flags (--dir/--host/--port) from the remaining
// args (command + its own flags/args). Only the known global flags are consumed;
// everything else — positionals AND unrecognized flags such as a subcommand's
// --priority/--tags/--json — is passed through verbatim so the subcommand can
// parse it. Global flags may appear anywhere, before or after the command.
func ParseTopLevel(args []string) (GlobalFlags, []string) {
	g := GlobalFlags{
		Dir:  os.Getenv("KANBAN_DIR"),
		Host: os.Getenv("HOST"),
		Port: os.Getenv("PORT"),
	}
	known := map[string]*string{"dir": &g.Dir, "host": &g.Host, "port": &g.Port}

	var rest []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			rest = append(rest, args[i:]...)
			break
		}
		if len(a) > 1 && a[0] == '-' {
			name := strings.TrimLeft(a, "-")
			val, inline := "", false
			if eq := strings.IndexByte(name, '='); eq >= 0 {
				val, name, inline = name[eq+1:], name[:eq], true
			}
			if ptr, ok := known[name]; ok {
				if inline {
					*ptr = val
				} else if i+1 < len(args) {
					*ptr = args[i+1]
					i++
				}
				continue
			}
		}
		rest = append(rest, a)
	}
	return g, rest
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
