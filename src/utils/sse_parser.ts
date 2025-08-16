export async function* parseSSEStream(
  reader: ReadableStreamDefaultReader<Uint8Array> | undefined,
): AsyncGenerator<string, void, unknown> {
  if (!reader) return;
  const decoder = new TextDecoder();
  let buf = "";

  try {
    while (true) {
      const { value, done } = await reader.read();
      if (done) break;
      if (value) {
        buf += decoder.decode(value, { stream: true });
      }

      let idx;
      while ((idx = buf.indexOf("\n\n")) !== -1) {
        const packet = buf.slice(0, idx).trim();
        buf = buf.slice(idx + 2);
        if (!packet) continue;
        for (const line of packet.split(/\r?\n/)) {
          const trimmed = line.trim();
          if (!trimmed.startsWith("data:")) continue;
          const payload = trimmed.slice(5).trim();
          if (payload === "[DONE]") return;
          try {
            const parsed = JSON.parse(payload);
            const delta = parsed?.choices?.[0]?.delta?.content ??
              parsed?.choices?.[0]?.message?.content;
            if (typeof delta === "string") yield delta;
          } catch {
            // ignore parse errors
          }
        }
      }
    }

    if (buf) {
      const packet = buf.trim();
      if (packet) {
        for (const line of packet.split(/\r?\n/)) {
          const trimmed = line.trim();
          if (!trimmed.startsWith("data:")) continue;
          const payload = trimmed.slice(5).trim();
          if (payload === "[DONE]") return;
          try {
            const parsed = JSON.parse(payload);
            const delta = parsed?.choices?.[0]?.delta?.content ??
              parsed?.choices?.[0]?.message?.content;
            if (typeof delta === "string") yield delta;
          } catch {
            // ignore
          }
        }
      }
    }
  } finally {
    // caller cancels reader; don't cancel here
  }
}
