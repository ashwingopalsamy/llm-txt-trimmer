# llmtrim

A tiny, deterministic text compactor for LLM input — available as a zero-cost browser UI and a Go CLI.

## Web UI

The web app runs entirely in the browser. Paste text on the left and the compacted result appears instantly on the right.

- no server
- no database
- no Docker
- no model/API calls
- pasted text never leaves the browser
- light and dark themes
- Safe, Compact and Dense modes
- copy and `.txt` download
- `.txt` / Markdown file input
- byte, whitespace and rough token-proxy savings
- responsive two-column → single-column layout

The static site lives in [`docs/`](./docs) so it can be hosted free with GitHub Pages.

### Publish with GitHub Pages

In the repository:

1. Open **Settings → Pages**.
2. Under **Build and deployment**, choose **Deploy from a branch**.
3. Select `main` and `/docs`.
4. Save.

The site will be available at:

```text
https://ashwingopalsamy.github.io/llm-txt-trimmer/
```

No runtime hosting or container is required.

## CLI

Install:

```sh
go install github.com/ashwingopalsamy/llm-txt-trimmer/cmd/llmtrim@latest
```

Use:

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

- Go CLI uses only the standard library
- web UI is plain HTML/CSS/JavaScript with no build system
- stdin/stdout friendly CLI
- UTF-8 / Tamil-safe
- preserves fenced code verbatim
- preserves YAML front matter verbatim
- preserves indentation on structural Markdown lines
- deterministic: same input + mode = same output

## Why not just `strings.Fields`?

Flattening all whitespace can destroy code, Markdown structure, nested lists and front matter. `llmtrim` compacts only where the transformation is predictable.

## Brand note

`llmtrim` is an independent project and is not affiliated with or endorsed by OpenAI. Its interface uses a neutral, minimal product-design language and does not use OpenAI trademarks or proprietary brand assets.
