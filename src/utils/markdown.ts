// Very small markdown renderer for terminal output.
// Supports headings (#..), bold (**text**), italic (*text*), inline code (`x`),
// code fences (```), and unordered lists (-, *, +).

const ANSI_BOLD = "\x1b[1m";
const ANSI_DIM = "\x1b[2m";
const ANSI_RESET = "\x1b[0m";
const ANSI_CYAN = "\x1b[36m";
const ANSI_YELLOW = "\x1b[33m";
const ANSI_GREEN = "\x1b[32m";

export function renderMarkdown(md: string): string {
  // Only process if markdown patterns are present
  const hasMarkdown =
    /(^#{1,6}\s+)|(^\s*[-*+]\s+)|(```)|(`[^`]+`)|(\*\*[^*]+\*\*)|(\*[^*]+\*)/
      .test(md);
  if (!hasMarkdown) return md;

  const lines = md.split(/\r?\n/);
  const out: string[] = [];
  let inCodeFence = false;

  for (const line of lines) {
    if (line.trim().startsWith("```")) {
      inCodeFence = !inCodeFence;
      if (inCodeFence) {
        out.push(ANSI_DIM + "--- code ---" + ANSI_RESET);
      } else {
        out.push(ANSI_DIM + "--- end code ---" + ANSI_RESET);
      }
      continue;
    }

    if (inCodeFence) {
      out.push(ANSI_DIM + line + ANSI_RESET);
      continue;
    }

    // Headings
    const h = line.match(/^(#{1,6})\s+(.*)$/);
    if (h) {
      const level = h[1].length;
      const text = h[2];
      const col = level <= 2 ? ANSI_CYAN : ANSI_YELLOW;
      out.push(`${ANSI_BOLD}${col}${text}${ANSI_RESET}`);
      continue;
    }

    // Unordered list
    const ul = line.match(/^\s*([-*+])\s+(.*)$/);
    if (ul) {
      out.push(`  • ${ul[2]}`);
      continue;
    }

    // Inline code `x`
    let processed = line.replace(
      /`([^`]+)`/g,
      (_m: string, p1: string) => `${ANSI_GREEN}${p1}${ANSI_RESET}`,
    );

    // Bold **text**
    processed = processed.replace(
      /\*\*([^*]+)\*\*/g,
      (_m: string, p1: string) => `${ANSI_BOLD}${p1}${ANSI_RESET}`,
    );
    // Italic *text*
    processed = processed.replace(
      /\*([^*]+)\*/g,
      (_m: string, p1: string) => `${ANSI_YELLOW}${p1}${ANSI_RESET}`,
    );

    out.push(processed);
  }

  return out.join("\n");
}
