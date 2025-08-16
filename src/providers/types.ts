export interface ChatMessage {
  role: "system" | "user" | "assistant";
  content: string;
}

export interface ChatRequest {
  model?: string;
  messages: ChatMessage[];
  stream?: boolean;
}

export type Fetcher = (input: string, init?: RequestInit) => Promise<Response>;

export type ProviderConfig = {
  model?: string;
  system?: string;
  prompt?: string;
  temperature?: number;
};

// Options that may be passed from the caller (CLI/core/tests) into provider
// adapters to avoid providers reading environment variables directly.
export type ProviderOptions = {
  apiKey?: string;
  fetcher?: Fetcher;
  baseUrl?: string;
};

// Minimal interface adapters should satisfy. This is for TypeScript typing
// and for runtime shape checks when dynamically importing adapters.
export interface ProviderAdapter {
  callProvider: (
    config: ProviderConfig,
    opts?: ProviderOptions,
  ) => Promise<{ text?: string; markdown?: string }>;
  chatCompletionStream?: (
    req: ChatRequest,
    opts?: ProviderOptions,
  ) => Promise<AsyncGenerator<string, void, unknown>>;
}
