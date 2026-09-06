(() => {
  'use strict';

  const MODES = new Set(['safe', 'compact', 'dense']);
  const encoder = new TextEncoder();

  function transform(input, mode = 'safe') {
    if (!MODES.has(mode)) throw new Error(`unknown mode ${JSON.stringify(mode)}`);

    const normalized = input.replace(/\r\n/g, '\n').replace(/\r/g, '\n');
    const lines = normalized.split('\n');
    const out = [];
    let inFence = false;
    let fenceChar = '';
    let fenceLen = 0;
    let inFrontMatter = false;
    let frontMatterSeen = false;
    let paragraph = [];

    const flushParagraph = () => {
      if (paragraph.length === 0) return;
      if (mode === 'safe') out.push(...paragraph);
      else out.push(joinParagraph(paragraph, mode === 'dense'));
      paragraph = [];
    };

    const appendBlank = () => {
      if (out.length === 0 || out[out.length - 1] === '') return;
      if (mode !== 'dense') out.push('');
    };

    lines.forEach((raw, index) => {
      const line = raw;
      const trimmed = line.trim();

      if (!frontMatterSeen && index === 0 && trimmed === '---') {
        flushParagraph();
        inFrontMatter = true;
        frontMatterSeen = true;
        out.push(line);
        return;
      }

      if (inFrontMatter) {
        out.push(line);
        if (trimmed === '---' || trimmed === '...') inFrontMatter = false;
        return;
      }

      const marker = fenceMarker(line);
      if (marker) {
        flushParagraph();
        if (!inFence) {
          inFence = true;
          fenceChar = marker.char;
          fenceLen = marker.length;
        } else if (marker.char === fenceChar && marker.length >= fenceLen) {
          inFence = false;
          fenceChar = '';
          fenceLen = 0;
        }
        out.push(line);
        return;
      }

      if (inFence) {
        out.push(line);
        return;
      }

      if (trimmed === '') {
        flushParagraph();
        appendBlank();
        return;
      }

      if (mode === 'safe') {
        paragraph.push(line);
        return;
      }

      if (isStructural(line)) {
        flushParagraph();
        out.push(normalizeStructural(line, mode === 'dense'));
        return;
      }

      paragraph.push(line);
    });

    flushParagraph();
    while (out.length && out[0] === '') out.shift();
    while (out.length && out[out.length - 1] === '') out.pop();

    const result = out.join('\n');
    return { text: result, stats: makeStats(input, result) };
  }

  function joinParagraph(lines, dense) {
    const parts = [];
    for (const line of lines) {
      let value = line.trim();
      if (dense) value = collapseWhitespace(value);
      if (value) parts.push(value);
    }
    return parts.join(' ');
  }

  function normalizeStructural(line, dense) {
    if (!dense) return line.replace(/[ \t]+$/g, '');
    const match = line.match(/^[ \t]*/);
    const prefix = match ? match[0] : '';
    const body = collapseWhitespace(line.slice(prefix.length).trim());
    return prefix + body;
  }

  function fenceMarker(line) {
    const leading = (line.match(/^ */) || [''])[0].length;
    if (leading > 3) return null;
    const rest = line.slice(leading);
    if (rest.length < 3) return null;
    const char = rest[0];
    if (char !== '`' && char !== '~') return null;
    let length = 0;
    while (rest[length] === char) length += 1;
    return length >= 3 ? { char, length } : null;
  }

  function isStructural(line) {
    const trimmed = line.trim();
    if (!trimmed) return false;

    const leading = (line.match(/^[ \t]*/) || [''])[0];
    if (leading.length >= 4 || line.startsWith('\t')) return true;

    return trimmed.startsWith('#') ||
      trimmed.startsWith('>') ||
      trimmed.startsWith('|') ||
      isHorizontalRule(trimmed) ||
      isListItem(trimmed) ||
      isDefinitionLike(trimmed) ||
      (trimmed.startsWith('<') && trimmed.endsWith('>'));
  }

  function isHorizontalRule(value) {
    const compact = value.replace(/[ \t]/g, '');
    if (compact.length < 3) return false;
    const char = compact[0];
    if (!['-', '*', '_'].includes(char)) return false;
    return Array.from(compact).every((item) => item === char);
  }

  function isListItem(value) {
    if (/^[-*+]\s/u.test(value)) return true;
    return /^\d+[.)]\s/u.test(value);
  }

  function isDefinitionLike(value) {
    const index = value.indexOf(':');
    return index > 0 && index <= 40 && !/[.!?]/.test(value.slice(0, index));
  }

  function collapseWhitespace(value) {
    return value.replace(/\s+/gu, ' ');
  }

  function countWhitespace(value) {
    const matches = value.match(/\s/gu);
    return matches ? matches.length : 0;
  }

  function lineCount(value) {
    if (!value) return 0;
    const normalized = value.replace(/\r\n/g, '\n').replace(/\r/g, '\n');
    return normalized.split('\n').length;
  }

  function runeCount(value) {
    return Array.from(value).length;
  }

  function tokenProxy(runes) {
    return runes === 0 ? 0 : Math.ceil(runes / 4);
  }

  function makeStats(before, after) {
    const beforeBytes = encoder.encode(before).length;
    const afterBytes = encoder.encode(after).length;
    const beforeRunes = runeCount(before);
    const afterRunes = runeCount(after);
    const savedBytes = beforeBytes - afterBytes;

    return {
      beforeBytes,
      afterBytes,
      savedBytes,
      savedPercent: beforeBytes === 0 ? 0 : (savedBytes * 100) / beforeBytes,
      beforeWhitespace: countWhitespace(before),
      afterWhitespace: countWhitespace(after),
      beforeLines: lineCount(before),
      afterLines: lineCount(after),
      tokenProxyBefore: tokenProxy(beforeRunes),
      tokenProxyAfter: tokenProxy(afterRunes),
    };
  }

  globalThis.LLMTrim = Object.freeze({ transform });
})();
