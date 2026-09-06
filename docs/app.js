(() => {
  'use strict';

  const root = document.documentElement;
  const input = document.getElementById('input-text');
  const output = document.getElementById('output-text');
  const inputCounter = document.getElementById('input-counter');
  const outputCounter = document.getElementById('output-counter');
  const clearButton = document.getElementById('clear-button');
  const copyButton = document.getElementById('copy-button');
  const downloadButton = document.getElementById('download-button');
  const fileInput = document.getElementById('file-input');
  const themeToggle = document.getElementById('theme-toggle');
  const modeButtons = Array.from(document.querySelectorAll('.mode-button'));
  const status = document.getElementById('status');
  const bytesSaved = document.getElementById('bytes-saved');
  const savedPercent = document.getElementById('saved-percent');
  const whitespaceChange = document.getElementById('whitespace-change');
  const tokenChange = document.getElementById('token-change');

  let mode = 'safe';
  let theme = readTheme();

  applyTheme(theme);

  function readTheme() {
    try {
      const stored = window.localStorage.getItem('llmtrim-theme');
      if (stored === 'light' || stored === 'dark') return stored;
    } catch {}
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }

  function rememberTheme(next) {
    try { window.localStorage.setItem('llmtrim-theme', next); } catch {}
  }

  function applyTheme(next) {
    theme = next;
    root.dataset.theme = next;
    const label = next === 'dark' ? 'Switch to light theme' : 'Switch to dark theme';
    themeToggle.setAttribute('aria-label', label);
    themeToggle.title = label;
  }

  themeToggle.addEventListener('click', () => {
    const next = theme === 'dark' ? 'light' : 'dark';
    rememberTheme(next);
    applyTheme(next);
  });

  function pluralizeCharacters(value) {
    const count = Array.from(value).length;
    return `${count.toLocaleString()} ${count === 1 ? 'character' : 'characters'}`;
  }

  function setStatus(message, isError = false) {
    status.textContent = message;
    status.classList.toggle('error', isError);
  }

  function setStats(stats = {}) {
    bytesSaved.textContent = Number(stats.savedBytes || 0).toLocaleString();
    savedPercent.textContent = `${Number(stats.savedPercent || 0).toFixed(1)}%`;
    whitespaceChange.textContent = `${Number(stats.beforeWhitespace || 0).toLocaleString()} → ${Number(stats.afterWhitespace || 0).toLocaleString()}`;
    tokenChange.textContent = `${Number(stats.tokenProxyBefore || 0).toLocaleString()} → ${Number(stats.tokenProxyAfter || 0).toLocaleString()}`;
  }

  function resetResult() {
    output.value = '';
    outputCounter.textContent = '0 characters';
    copyButton.disabled = true;
    downloadButton.disabled = true;
    setStats();
    setStatus('Ready');
  }

  function runTrim() {
    inputCounter.textContent = pluralizeCharacters(input.value);
    if (!input.value) {
      resetResult();
      return;
    }

    try {
      const result = globalThis.LLMTrim.transform(input.value, mode);
      output.value = result.text;
      outputCounter.textContent = pluralizeCharacters(output.value);
      copyButton.disabled = output.value.length === 0;
      downloadButton.disabled = output.value.length === 0;
      setStats(result.stats);
      setStatus('Updated locally');
    } catch (error) {
      setStatus(error instanceof Error ? error.message : 'Could not trim text', true);
    }
  }

  input.addEventListener('input', runTrim);

  modeButtons.forEach((button) => {
    button.addEventListener('click', () => {
      mode = button.dataset.mode || 'safe';
      modeButtons.forEach((candidate) => {
        const selected = candidate === button;
        candidate.classList.toggle('active', selected);
        candidate.setAttribute('aria-pressed', selected ? 'true' : 'false');
      });
      runTrim();
    });
  });

  clearButton.addEventListener('click', () => {
    input.value = '';
    inputCounter.textContent = '0 characters';
    fileInput.value = '';
    resetResult();
    input.focus();
  });

  fileInput.addEventListener('change', async () => {
    const file = fileInput.files && fileInput.files[0];
    if (!file) return;
    if (file.size > 8 * 1024 * 1024) {
      setStatus('File is larger than the 8 MiB browser limit', true);
      fileInput.value = '';
      return;
    }

    try {
      input.value = await file.text();
      runTrim();
      input.focus();
    } catch {
      setStatus('Could not read that file', true);
    }
  });

  copyButton.addEventListener('click', async () => {
    if (!output.value) return;

    try {
      await navigator.clipboard.writeText(output.value);
      setStatus('Copied');
      return;
    } catch {}

    output.focus();
    output.select();
    const copied = document.execCommand('copy');
    setStatus(copied ? 'Copied' : 'Copy failed', !copied);
  });

  downloadButton.addEventListener('click', () => {
    if (!output.value) return;
    const blob = new Blob([output.value], { type: 'text/plain;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement('a');
    anchor.href = url;
    anchor.download = 'trimmed.txt';
    document.body.appendChild(anchor);
    anchor.click();
    anchor.remove();
    window.setTimeout(() => URL.revokeObjectURL(url), 0);
    setStatus('Downloaded');
  });
})();
