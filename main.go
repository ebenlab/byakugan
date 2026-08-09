// Command byakugan serves a folder of HTML architecture records — ADRs,
// system overviews, design decisions — as a navigable, searchable site.
// Other HTML docs (PRDs, runbooks) are served just as well, but the tool is
// built architecture-first.
//
// Usage:
//
//	byakugan [flags] [folder]
//	byakugan <subcommand>
//
// The folder defaults to the current directory. Every first-level
// subdirectory is treated as a project; every HTML file inside it becomes a
// searchable, navigable page.
//
// The subcommands (help, version, style, rules, template) are agent-facing:
// they print the shared doc stylesheet, the authoring rules, and doc
// generation prompt templates so coding agents can produce pages that match
// the design system.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ebenlab/byakugan/internal/agentkit"
	"github.com/ebenlab/byakugan/internal/index"
	"github.com/ebenlab/byakugan/internal/selfupdate"
	"github.com/ebenlab/byakugan/internal/server"
	"github.com/ebenlab/byakugan/internal/watcher"
)

// version is stamped at release time via -ldflags "-X main.version=v1.2.3".
var version = "dev"

func main() {
	log.SetFlags(0)

	if handled, code := run(os.Args[1:], os.Stdout, os.Stderr); handled {
		os.Exit(code)
	}
	serve(os.Args[1:])
}

// run dispatches the agent-facing subcommands: help, version, style, rules,
// and template. It reports handled=false when the first argument names no
// subcommand, in which case the caller falls through to serve mode — so
// `byakugan [flags] [folder]` keeps working unchanged.
func run(args []string, stdout, stderr io.Writer) (handled bool, code int) {
	if len(args) == 0 {
		return false, 0
	}
	switch args[0] {
	case "help":
		fs, _ := newServeFlags(stdout)
		printUsage(stdout, fs)
		return true, 0
	case "version":
		fmt.Fprintln(stdout, "byakugan", version)
		return true, 0
	case "upgrade":
		if err := selfupdate.New().Upgrade(context.Background(), version, stdout); err != nil {
			fmt.Fprintf(stderr, "byakugan: %v\n", err)
			return true, 1
		}
		return true, 0
	case "style":
		stdout.Write(agentkit.StyleCSS())
		return true, 0
	case "rules":
		stdout.Write(agentkit.Rules())
		return true, 0
	case "template":
		if len(args) < 2 {
			fmt.Fprintf(stderr, "byakugan template <kind> — available kinds: %s\n", strings.Join(agentkit.Kinds(), ", "))
			return true, 2
		}
		tpl, err := agentkit.Template(args[1])
		if err != nil {
			fmt.Fprintf(stderr, "byakugan: %v — available kinds: %s\n", err, strings.Join(agentkit.Kinds(), ", "))
			return true, 2
		}
		stdout.Write(tpl)
		return true, 0
	}
	return false, 0
}

// serveOptions holds the parsed serve-mode flags.
type serveOptions struct {
	port     int
	host     string
	open     bool
	noWatch  bool
	noUpdate bool
	version  bool
}

// newServeFlags defines the serve-mode flag set. Usage and parse errors are
// written to w.
func newServeFlags(w io.Writer) (*flag.FlagSet, *serveOptions) {
	opts := &serveOptions{}
	fs := flag.NewFlagSet("byakugan", flag.ExitOnError)
	fs.SetOutput(w)
	fs.IntVar(&opts.port, "port", 4664, "port to listen on")
	fs.StringVar(&opts.host, "host", "127.0.0.1", "interface to bind (use 0.0.0.0 to expose on the network)")
	fs.BoolVar(&opts.open, "open", false, "open the site in the default browser after starting")
	fs.BoolVar(&opts.noWatch, "no-watch", false, "disable file watching and live reload")
	fs.BoolVar(&opts.noUpdate, "no-update-check", false, "disable the startup check for a newer release")
	fs.BoolVar(&opts.version, "version", false, "print version and exit")
	fs.Usage = func() { printUsage(w, fs) }
	return fs, opts
}

// printUsage writes the full usage text — serve flags plus subcommands — to w.
func printUsage(w io.Writer, fs *flag.FlagSet) {
	fmt.Fprintf(w, `byakugan — a tiny live server for architecture records (ADRs) in HTML

Usage:
  byakugan [flags] [folder]   serve a docs folder (default ".")
  byakugan <subcommand>

Subcommands:
  help             print this help
  version          print version and exit
  style            print the shared doc stylesheet (doc.css) to stdout
  rules            print the doc authoring guide for agents
  template <kind>  print a doc generation prompt template
                   kinds: %s
  upgrade          replace this binary with the latest release

Flags:
`, strings.Join(agentkit.Kinds(), ", "))
	fs.SetOutput(w)
	fs.PrintDefaults()
	fmt.Fprintf(w, `
Examples:
  byakugan ./docs
  byakugan --port 8080 --open ~/team/architecture
  byakugan style > docs/_shared/doc.css
  byakugan template adr
`)
}

// serve runs the live server: parse flags, index the folder, watch it, and
// listen until interrupted.
func serve(args []string) {
	fs, opts := newServeFlags(os.Stderr)
	// ExitOnError: Parse exits the process on a bad flag, so the returned
	// error is always nil here.
	_ = fs.Parse(args)

	if opts.version {
		fmt.Println("byakugan", version)
		return
	}

	root := "."
	if fs.NArg() > 0 {
		root = fs.Arg(0)
	}
	root, err := filepath.Abs(root)
	if err != nil {
		log.Fatalf("byakugan: %v", err)
	}
	info, err := os.Stat(root)
	if err != nil {
		log.Fatalf("byakugan: %v", err)
	}
	if !info.IsDir() {
		log.Fatalf("byakugan: %s is not a directory", root)
	}

	idx := index.New(root)
	if err := idx.Rebuild(); err != nil {
		log.Fatalf("byakugan: indexing failed: %v", err)
	}

	srv := server.New(idx, version)

	if !opts.noWatch {
		w, err := watcher.New(root, func() {
			if err := idx.Rebuild(); err != nil {
				log.Printf("byakugan: re-index failed: %v", err)
				return
			}
			srv.Broadcast("reload")
		})
		if err != nil {
			log.Printf("byakugan: file watching unavailable: %v", err)
		} else {
			defer w.Close()
		}
	}

	addr := fmt.Sprintf("%s:%d", opts.host, opts.port)
	url := fmt.Sprintf("http://%s", addr)
	log.Printf("byakugan %s — serving %s", version, root)
	log.Printf("→ %s", url)

	if !opts.noUpdate {
		// Non-blocking; silent when offline or already current. Opt out with
		// --no-update-check or BYAKUGAN_NO_UPDATE_CHECK=1.
		selfupdate.Notice(version, func(tag string) {
			log.Printf("↑ %s is available (you have %s) — run 'byakugan upgrade'", tag, version)
		})
	}

	if opts.open {
		go openBrowser(url)
	}
	if err := srv.ListenAndServe(addr); err != nil {
		log.Fatalf("byakugan: %v", err)
	}
}

// openBrowser launches the platform's default browser at url. Failures are
// non-fatal: the URL is already printed to the terminal.
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("byakugan: could not open browser: %v", err)
	}
}
