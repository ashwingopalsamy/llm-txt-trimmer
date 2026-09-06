package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ashwingopalsamy/llm-txt-trimmer/internal/compact"
	"github.com/ashwingopalsamy/llm-txt-trimmer/internal/webapp"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "serve" {
		runServer(os.Args[2:])
		return
	}
	runCLI(os.Args[1:])
}

func runCLI(args []string) {
	flags := flag.NewFlagSet("llmtrim", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	modeFlag := flags.String("mode", "safe", "compaction mode: safe, compact, or dense")
	outPath := flags.String("o", "", "write output to file instead of stdout")
	stats := flags.Bool("stats", false, "print size statistics to stderr")
	flags.Usage = func() {
		fmt.Fprintf(flags.Output(), "Usage: llmtrim [flags] [file]\n       llmtrim serve [flags]\n\nReads stdin when no file is given.\n\n")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fatalf("%v", err)
	}

	if flags.NArg() > 1 {
		fatalf("expected at most one input file")
	}

	input, err := readInput(flags.Args())
	if err != nil {
		fatalf("read input: %v", err)
	}

	output, s, err := compact.Transform(string(input), compact.Mode(*modeFlag))
	if err != nil {
		fatalf("%v", err)
	}

	if *outPath == "" {
		if _, err := io.WriteString(os.Stdout, output); err != nil {
			fatalf("write output: %v", err)
		}
	} else if err := os.WriteFile(*outPath, []byte(output), 0o644); err != nil {
		fatalf("write %s: %v", *outPath, err)
	}

	if *stats {
		fmt.Fprintf(os.Stderr,
			"\nbytes: %d -> %d (%d saved, %.1f%%)\nwhitespace: %d -> %d\nlines: %d -> %d\ntoken proxy*: %d -> %d\n* rough chars/4 heuristic; actual tokenizer counts vary by model and language\n",
			s.BeforeBytes, s.AfterBytes, s.SavedBytes(), s.SavedPercent(),
			s.BeforeWhitespace, s.AfterWhitespace,
			s.BeforeLines, s.AfterLines,
			s.TokenProxyBefore(), s.TokenProxyAfter(),
		)
	}
}

func runServer(args []string) {
	flags := flag.NewFlagSet("llmtrim serve", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	addr := flags.String("addr", defaultAddr(), "HTTP listen address")
	flags.Usage = func() {
		fmt.Fprintf(flags.Output(), "Usage: llmtrim serve [flags]\n\n")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fatalf("%v", err)
	}
	if flags.NArg() != 0 {
		fatalf("serve does not accept positional arguments")
	}

	server := &http.Server{
		Addr:              *addr,
		Handler:           webapp.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	fmt.Fprintf(os.Stderr, "llmtrim: web UI listening on http://%s\n", displayAddr(*addr))
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fatalf("serve: %v", err)
	}
}

func defaultAddr() string {
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}
	return "127.0.0.1:8080"
}

func displayAddr(addr string) string {
	if len(addr) > 0 && addr[0] == ':' {
		return "localhost" + addr
	}
	return addr
}

func readInput(args []string) ([]byte, error) {
	if len(args) == 0 || args[0] == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(args[0])
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "llmtrim: "+format+"\n", args...)
	os.Exit(2)
}
