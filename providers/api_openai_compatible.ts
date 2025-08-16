export type ChatMessage = {
  role: "system" | "user" | "assistant";
  content: string;
};
export type ChatRequest = {
  model?: string;
  messages: ChatMessage[];
  stream?: boolean;
};
export type Fetcher = (input: string, init?: RequestInit) => Promise<Response>;

export async function chatCompletionRequest(
  params: {
    url: string;
    apiKey: string;
    body: ChatRequest;
    fetcher?: Fetcher;
  },
): Promise<string> {
  const { url, apiKey, body, fetcher } = params;
  const realFetch: Fetcher = fetcher ?? ((input, init) => fetch(input, init));

  const res = await realFetch(url, {
    method: "POST",
    headers: {
      "Authorization": `Bearer ${apiKey}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(`provider error: ${res.status} ${res.statusText}: ${text}`);
  }

  const data = await res.json();
  const content = data?.choices?.[0]?.message?.content;
  if (typeof content !== "string") {
    throw new Error("invalid response shape from provider");
  }
  return content;
}

// Streaming variant: returns an async generator that yields string chunks as they arrive
export async function* chatCompletionRequestStream(
  params: {
    url: string;
    apiKey: string;
    body: ChatRequest;
    fetcher?: Fetcher;
  },
): AsyncGenerator<string, void, unknown> {
  const { url, apiKey, body, fetcher } = params;
  const realFetch: Fetcher = fetcher ?? ((input, init) => fetch(input, init));

  const res = await realFetch(url, {
    method: "POST",
    headers: {
      "Authorization": `Bearer ${apiKey}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(`provider error: ${res.status} ${res.statusText}: ${text}`);
  }

  const reader = res.body?.getReader();
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
        // handle multiple lines; process lines starting with `data:`
        for (const line of packet.split(/\r?\n/)) {
          const trimmed = line.trim();
          if (!trimmed.startsWith("data:")) continue;
          const payload = trimmed.slice(5).trim();
          if (payload === "[DONE]") return;
          try {
            const parsed = JSON.parse(payload);
            // accept both delta content (streaming) and message.content (non-streamed)
            const delta = parsed?.choices?.[0]?.delta?.content ??
              parsed?.choices?.[0]?.message?.content;
            if (typeof delta === "string") yield delta;
          } catch {
            // ignore parse errors
          }
        }
      }
    }

    // flush remaining buffer
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
    try {
      await reader.cancel();
    } catch {
      // ignore
    }
  }
}
