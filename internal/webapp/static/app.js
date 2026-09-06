(() => {
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
  let timer = null;
  let request = null;
  let theme = readTheme();

  applyTheme(theme);

  function readTheme() {
    const stored = window.localStorage.getItem('llmtrim-theme');
    if (stored === 'light' || stored === 'dark') return stored;
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }

  function applyTheme(next) {
    theme = next;
    root.dataset.theme = next;
    themeToggle.setAttribute('aria-label', next === 'dark' ? 'Switch to light theme' : 'Switch to dark theme');
    themeToggle.title = next === 'dark' ? 'Switch to light theme' : 'Switch to dark theme';
  }

  themeToggle.addEventListener('click', () => {
    const next = theme === 'dark' ? 'light' : 'dark';
    window.localStorage.setItem('llmtrim-theme', next);
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

  function setStats(stats) {
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
    setStats({});
    setStatus('Ready');
  }

  function scheduleTrim() {
    inputCounter.textContent = pluralizeCharacters(input.value);
    window.clearTimeout(timer);
    if (!input.value) {
      if (request) request.abort();
      resetResult();
      return;
    }
    setStatus('Trimming…');
    timer = window.setTimeout(runTrim, 110);
  }

  async function runTrim() {
    if (request) request.abort();
    request = new AbortController();

    try {
      const response = await fetch('/api/trim', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ text: input.value, mode }),
        signal: request.signal,
      });

      const body = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(body.error || `Request failed (${response.status})`);

      output.value = body.text;
      outputCounter.textContent = pluralizeCharacters(output.value);
      copyButton.disabled = output.value.length === 0;
      downloadButton.disabled = output.value.length === 0;
      setStats(body.stats || {});
      setStatus('Updated');
    } catch (error) {
      if (error.name === 'AbortError') return;
      setStatus(error.message || 'Could not trim text', true);
    }
  }

  input.addEventListener('input', scheduleTrim);

  modeButtons.forEach((button) => {
    button.addEventListener('click', () => {
      mode = button.dataset.mode;
      modeButtons.forEach((candidate) => {
        const selected = candidate === button;
        candidate.classList.toggle('active', selected);
        candidate.setAttribute('aria-pressed', selected ? 'true' : 'false');
      });
      scheduleTrim();
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
    if (file.size > 4 * 1024 * 1024) {
      setStatus('File is larger than the 4 MiB limit', true);
      fileInput.value = '';
      return;
    }
    try {
      input.value = await file.text();
      scheduleTrim();
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
    } catch {
      output.focus();
      output.select();
      const copied = document.execCommand('copy');
      setStatus(copied ? 'Copied' : 'Copy failed', !copied);
    }
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
