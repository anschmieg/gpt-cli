#!/usr/bin/env -S deno run --allow-net
/// <reference lib="deno.ns" />
/// <reference lib="dom" />

const port = 8086;

const customResponseText = `
__Advertisement :)__

- __[pica](https://nodeca.github.io/pica/demo/)__ - high quality and fast image
  resize in browser.

You will like those projects!

---

# h1 Heading 8-)
## h2 Heading
### h3 Heading
#### h4 Heading
##### h5 Heading
###### h6 Heading

## Horizontal Rules

___

---

***

## Typographic replacements

Enable typographer option to see result.

(c) (C) (r) (R) (tm) (TM) (p) (P) +- 

test.. test... test..... test?..... test!....

!!!!!! ???? ,,  -- ---

"Smartypants, double quotes" and 'single quotes'
`;

function buildResponseTemplate() {
  return {
    choices: [
      {
        message: { content: customResponseText, role: "assistant" },
        finish_reason: "stop",
        index: 0,
        logprobs: null,
      },
    ],
    id: "chatcmpl-mock",
    model: "gpt-3.5-mock",
    object: "chat.completion",
    usage: { completion_tokens: 10, prompt_tokens: 9, total_tokens: 19 },
    created: Math.floor(Date.now() / 1000),
  };
}

async function handler(req: Request): Promise<Response> {
  const url = new URL(req.url);

  if (url.pathname === "/health" && req.method === "GET") {
    return new Response(JSON.stringify({ status: "ok" }), {
      status: 200,
      headers: { "Content-Type": "application/json" },
    });
  }

  if (url.pathname === "/v1/chat/completions" && req.method === "POST") {
    // parse body as unknown and narrow to a plain object map; avoid `any`
    const rawBody: unknown = await req.json().catch(() => ({}));
    const body: Record<string, unknown> = (
        typeof rawBody === "object" && rawBody !== null
      )
      ? (rawBody as Record<string, unknown>)
      : {};
    const responseTemplate = buildResponseTemplate();

    // If client asked for streaming, return SSE
    if ("stream" in body && Boolean(body["stream"])) {
      const content = responseTemplate.choices?.[0]?.message?.content || "";
      const words = content.split(/\s+/).filter(Boolean);
      const batchSize = 8;
      const latencyMs = 3; // very small delay for CI speed

      const stream = new ReadableStream({
        async start(controller) {
          for (let i = 0; i < words.length; i += batchSize) {
            const chunk = words.slice(i, i + batchSize).join(" ") +
              (i + batchSize < words.length ? " " : "");
            const payload = JSON.stringify({
              choices: [{ delta: { content: chunk } }],
            });
            controller.enqueue(
              new TextEncoder().encode(`data: ${payload}\n\n`),
            );
            // small delay
            await new Promise((r) => setTimeout(r, latencyMs));
          }
          controller.enqueue(new TextEncoder().encode("data: [DONE]\n\n"));
          controller.close();
        },
      });

      return new Response(stream, {
        status: 200,
        headers: { "Content-Type": "text/event-stream" },
      });
    }

    // Non-streaming: return JSON immediately
    return new Response(JSON.stringify(responseTemplate), {
      status: 200,
      headers: { "Content-Type": "application/json" },
    });
  }

  return new Response("Not Found", { status: 404 });
}

export async function startMockServerProcess(): Promise<{ close: () => void }> {
  // Prefer spawning a subprocess when run permission is granted. Otherwise,
  // if the test runner granted network permission, start the server
  // in-process using Deno.serve and return a handle to stop it.
  try {
    const runPerm = await Deno.permissions.query({ name: "run" });
    if (runPerm.state === "granted") {
      const command = new Deno.Command(Deno.execPath(), {
        args: ["run", "--allow-net=127.0.0.1:8086", import.meta.url],
        stdout: "piped",
        stderr: "piped",
      });
      const proc = command.spawn();

      // Wait until the server is healthy or timeout
      const start = Date.now();
      while (Date.now() - start < 2000) {
        try {
          const res = await fetch("http://127.0.0.1:8086/health");
          try {
            await res.text();
          } catch {
            // ignore
          }
          if (res.ok) break;
        } catch (_e) {
          // server not up yet
        }
        await new Promise((r) => setTimeout(r, 10));
      }

      return {
        close: async () => {
          try {
            proc.kill();
          } catch (_e) {
            // ignore kill errors
          }

          // Close or cancel stdout/stderr if available to avoid leaks reported by
          // the Deno test runner. Different Deno versions expose slightly
          // different stream APIs, so try both cancel() and close() where
          // present.
          try {
            type StreamLike = {
              cancel?: () => Promise<void> | void;
              close?: () => void;
            };

            if (proc.stdout) {
              try {
                const s = proc.stdout as unknown as StreamLike;
                if (typeof s.cancel === "function") {
                  await s.cancel();
                }
              } catch {
                // ignore
              }
              try {
                const s = proc.stdout as unknown as StreamLike;
                if (typeof s.close === "function") {
                  s.close();
                }
              } catch {
                // ignore
              }
            }
            if (proc.stderr) {
              try {
                const s = proc.stderr as unknown as StreamLike;
                if (typeof s.cancel === "function") {
                  await s.cancel();
                }
              } catch {
                // ignore
              }
              try {
                const s = proc.stderr as unknown as StreamLike;
                if (typeof s.close === "function") {
                  s.close();
                }
              } catch {
                // ignore
              }
            }
          } catch {
            // ignore any close/cancel errors
          }

          try {
            // Wait for process to exit
            return await proc.status.catch(() => {});
          } catch (_e) {
            // ignore
            return Promise.resolve();
          }
        },
      };
    }
  } catch {
    // ignore permission query failures and fall through to net-based attempt
  }

  // If run isn't granted, try to start the server in-process when network
  // permission is available. This avoids requiring --allow-run for tests that
  // are allowed to bind to localhost.
  try {
    const netPerm = await Deno.permissions.query({
      name: "net",
      host: "127.0.0.1:8086",
    });
    if (netPerm.state === "granted") {
      const controller = new AbortController();
      // Start server in background; stop by aborting the controller
      Deno.serve(
        { hostname: "127.0.0.1", port, signal: controller.signal },
        handler,
      );

      // Wait for quick health check readiness
      const start = Date.now();
      while (Date.now() - start < 2000) {
        try {
          const res = await fetch("http://127.0.0.1:8086/health");
          if (res.ok) break;
        } catch {
          // not up yet
        }
        await new Promise((r) => setTimeout(r, 10));
      }

      return {
        close: () => {
          try {
            controller.abort();
          } catch {
            // ignore
          }
          return Promise.resolve();
        },
      };
    }
  } catch {
    // ignore permission query failures
  }

  throw new Error(
    "startMockServerProcess requires --allow-run or --allow-net=127.0.0.1:8086 to start the mock server",
  );
}

console.log(`Mock OpenAI API server module loaded at http://127.0.0.1:${port}`);
// If executed directly, start the server in-process.
if (import.meta.main) {
  console.log(`Mock OpenAI API server listening at http://127.0.0.1:${port}`);
  // Use the built-in Deno.serve to avoid deprecated std APIs.
  // Deno.serve will block the process and handle incoming requests.
  Deno.serve({ hostname: "127.0.0.1", port }, handler);
}
