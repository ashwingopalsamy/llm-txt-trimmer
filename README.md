# llmtrim

A tiny, deterministic text compactor for LLM input — available as both a browser UI and a CLI.

Paste text in the web app and `llmtrim` removes formatting waste while preserving code fences, YAML front matter, Markdown structure and UTF-8 text. No model call is involved.

## Web UI

Run locally:

```sh
go run ./cmd/llmtrim serve
```

Then open `http://localhost:8080`.

Or install once:

```sh
go install github.com/ashwingopalsamy/llm-txt-trimmer/cmd/llmtrim@latest
llmtrim serve
```

The UI provides:

- live two-column original → trimmed editing
- Safe, Compact and Dense modes
- light and dark themes
- copy and `.txt` download actions
- `.txt` / Markdown file input
- byte, whitespace and rough token-proxy savings
- responsive single-column layout on narrow screens
- no persistence or database

The web server uses the same Go compaction engine as the CLI, so browser and CLI output stay consistent.

### Deploy

`llmtrim serve` honors the `PORT` environment variable automatically. A minimal multi-stage Docker image is included:

```sh
docker build -t llmtrim .
docker run --rm -p 8080:8080 llmtrim
```

Health check:

```text
GET /healthz
```

The trim endpoint accepts up to 4 MiB per request and does not persist submitted text.

## CLI

```sh
cat prompt.md | llmtrim
llmtrim --mode compact prompt.md
llmtrim --mode dense --stats prompt.md > prompt.min.md
llmtrim --mode compact -o prompt.min.md prompt.md
```

### Modes

- `safe` — trims outer whitespace-only space and collapses repeated blank lines. It does not rewrite non-empty lines.
- `compact` — additionally joins wrapped prose lines while keeping Markdown-like structural lines separate.
- `dense` — does the above and collapses repeated whitespace outside fenced code/front matter.

`--stats` reports bytes, whitespace, line count, and a clearly labelled `chars/4` token proxy. Real token counts vary by tokenizer, model and language; `llmtrim` does not claim the proxy is an exact token count.

## Design constraints

- Go standard library only
- one executable for CLI + web server
- no frontend build system
- embedded HTML/CSS/JavaScript
- stdin/stdout friendly
- UTF-8 / Tamil-safe
- preserves fenced code verbatim
- preserves YAML front matter verbatim
- preserves indentation on structural Markdown lines
- deterministic: same input + mode = same output

## Why not just `strings.Fields`?

Flattening all whitespace can destroy code, Markdown structure, nested lists and front matter. `llmtrim` compacts only where the transformation is predictable.

## Brand note

`llmtrim` is an independent project and is not affiliated with or endorsed by OpenAI. Its interface uses a neutral, minimal product-design language and does not use OpenAI trademarks or proprietary brand assets.
