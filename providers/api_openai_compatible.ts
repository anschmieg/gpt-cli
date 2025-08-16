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
