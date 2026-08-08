// Command byakugan serves a folder of agent-generated architecture and PRD
// HTML documents as a navigable, searchable site.
//
// Usage:
//
//	byakugan [flags] [folder]
//
// The folder defaults to the current directory. Every first-level
// subdirectory is treated as a project; every HTML file inside it becomes a
// searchable, navigable page.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/ebenlab/byakugan/internal/index"
	"github.com/ebenlab/byakugan/internal/server"
	"github.com/ebenlab/byakugan/internal/watcher"
)

// version is stamped at release time via -ldflags "-X main.version=v1.2.3".
var version = "dev"

func main() {
	log.SetFlags(0)

	port := flag.Int("port", 4664, "port to listen on")
	host := flag.String("host", "127.0.0.1", "interface to bind (use 0.0.0.0 to expose on the network)")
	open := flag.Bool("open", false, "open the site in the default browser after starting")
	noWatch := flag.Bool("no-watch", false, "disable file watching and live reload")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Usage = usage
	flag.Parse()

	if *showVersion {
		fmt.Println("byakugan", version)
		return
	}

	root := "."
	if flag.NArg() > 0 {
		root = flag.Arg(0)
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

	if !*noWatch {
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

	addr := fmt.Sprintf("%s:%d", *host, *port)
	url := fmt.Sprintf("http://%s", addr)
	log.Printf("byakugan %s — serving %s", version, root)
	log.Printf("→ %s", url)

	if *open {
		go openBrowser(url)
	}
	if err := srv.ListenAndServe(addr); err != nil {
		log.Fatalf("byakugan: %v", err)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `byakugan — a tiny live server for architecture docs and PRDs

Usage:
  byakugan [flags] [folder]

Flags:
`)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, `
Examples:
  byakugan ./docs
  byakugan --port 8080 --open ~/team/architecture
`)
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
