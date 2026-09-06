# llmtrim

Tiny, deterministic text compactor for LLM input.

It removes formatting waste without calling an AI model. The default mode is deliberately conservative and fenced code/front matter are preserved.

## Install

```sh
go install github.com/ashwingopalsamy/llm-txt-trimmer/cmd/llmtrim@latest
```

## Use

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

- no runtime dependencies
- stdin/stdout friendly
- UTF-8 / Tamil-safe
- preserves fenced code verbatim
- preserves YAML front matter verbatim
- preserves indentation on structural Markdown lines
- deterministic: same input + mode = same output

## Why not just `strings.Fields`?

Flattening all whitespace can destroy code, Markdown structure, nested lists and front matter. `llmtrim` compacts only where the transformation is predictable.
