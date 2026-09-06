package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/ashwingopalsamy/llm-txt-trimmer/internal/compact"
)

func main() {
	var (
		modeFlag = flag.String("mode", "safe", "compaction mode: safe, compact, or dense")
		outPath  = flag.String("o", "", "write output to file instead of stdout")
		stats    = flag.Bool("stats", false, "print size statistics to stderr")
	)
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: llmtrim [flags] [file]\n\nReads stdin when no file is given.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() > 1 {
		fatalf("expected at most one input file")
	}

	input, err := readInput(flag.Args())
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
