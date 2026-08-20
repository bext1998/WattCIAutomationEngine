function copyText(text) {
  if (navigator.clipboard && window.isSecureContext) {
    return navigator.clipboard.writeText(text);
  }

  const fallback = document.createElement('textarea');
  fallback.value = text;
  fallback.setAttribute('readonly', '');
  fallback.style.position = 'fixed';
  fallback.style.opacity = '0';
  document.body.append(fallback);
  fallback.select();
  const copied = document.execCommand('copy');
  fallback.remove();
  return copied ? Promise.resolve() : Promise.reject(new Error('copy failed'));
}

function copyCode(button) {
  const code = button.closest('.code-block').querySelector('pre').innerText;
  const original = button.textContent;
  button.textContent = '已複製';
  copyText(code).catch(() => { button.textContent = '請手動複製'; });
  window.setTimeout(() => { button.textContent = original; }, 1600);
}

function startTerminal() {
  const terminal = document.querySelector('[data-terminal]');
  if (!terminal) return;
  const output = terminal.querySelector('[data-terminal-output]');
  const text = terminal.querySelector('template').content.textContent.trim();
  if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) { output.textContent = text; return; }
  let index = 0;
  const write = () => {
    output.textContent += text[index] ?? '';
    index += 1;
    if (index < text.length) window.setTimeout(write, text[index - 1] === '\n' ? 180 : 17);
  };
  write();
}

document.querySelectorAll('[data-copy]').forEach((button) => button.addEventListener('click', () => copyCode(button)));
startTerminal();
