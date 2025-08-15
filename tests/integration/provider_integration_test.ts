import { chatCompletion } from "../../src/providers/openai.ts";

// helper: run a promise with a timeout and ensure timer is cleared to avoid leaks
async function withTimeout<T>(p: Promise<T>, ms: number) {
  let timer: number | undefined;
  const timeout = new Promise<never>((_res, rej) => {
    timer = setTimeout(
      () => rej(new Error("provider call timed out")),
      ms,
    ) as unknown as number;
  });
  try {
    return await Promise.race([p, timeout]);
  } finally {
    if (timer) clearTimeout(timer);
  }
}

async function startMockServer() {
  // First try starting the Deno-based mock server using the current Deno executable.
  const denoCmd = new Deno.Command(Deno.execPath(), {
    args: [
      "run",
      "--allow-net=127.0.0.1:8086",
      "--no-check",
      "./mock-openai/mock-server.ts",
    ],
    cwd: ".",
    stdout: "null",
    stderr: "null",
  });
  try {
    const child = denoCmd.spawn();
    // Poll /health until ready
    const base = "http://127.0.0.1:8086";
    const deadline = Date.now() + 8000;
    while (Date.now() < deadline) {
      try {
        const res = await fetch(`${base}/health`);
        // consume body to avoid resource leaks
        try {
          await res.text();
        } catch {
          // ignore
        }
        if (res.ok) return child;
      } catch {
        // ignore
      }
      await new Promise((r) => setTimeout(r, 50));
    }
    // didn't start in time
    try {
      child.kill();
    } catch {
      // ignore
    }
  } catch {
    // Deno runner not available or failed; fall through to other runners below
  }

  // Fall back: try node/bun runners in the mock-openai folder.
  const runners = [
    { cmd: "bun", args: ["run", "index.ts"] },
    { cmd: "node", args: ["./mock-server.js"] },
  ];

  for (const r of runners) {
    try {
      const command = new Deno.Command(r.cmd, {
        args: r.args,
        cwd: "./mock-openai",
        stdout: "null",
        stderr: "null",
      });
      const child = command.spawn();

      // Poll /health until ready
      const deadline = Date.now() + 10_000;
      while (Date.now() < deadline) {
        try {
          const res = await fetch("http://127.0.0.1:8086/health");
          try {
            await res.text();
          } catch {
            // ignore
          }
          if (res.ok) return child;
        } catch {
          // ignore
        }
        await new Promise((r) => setTimeout(r, 100));
      }

      try {
        child.kill();
        await child.status;
      } catch {
        // ignore
      }
    } catch {
      // runner not available; try next
    }
  }

  throw new Error("mock server failed to start with any runner");
}

// removed stream-reading readiness logic; using HTTP polling instead

Deno.test("provider chatCompletion hits mock-openai and returns markdown (with server)", async () => {
  // Start the mock server and ensure it's killed after the test.
  const server = await startMockServer();
  try {
    // Set test env so provider refuses external endpoints
    Deno.env.set("GPT_CLI_TEST", "1");

    const req = {
      messages: [
        { role: "user" as const, content: "Say something markdowny" },
      ],
    };

    // Use top-level withTimeout helper to avoid timer/resource leaks in tests.
    const content = await withTimeout(
      chatCompletion(req, "http://127.0.0.1:8086"),
      5000,
    );
    if (!content.includes("Advertisement") && !content.includes("h1 Heading")) {
      throw new Error(
        "unexpected provider content: missing expected markdown snippets",
      );
    }
  } finally {
    try {
      server.kill();
      // Wait for the process to actually exit
      try {
        await server.status;
      } catch {
        // ignore
      }
    } catch {
      // ignore
    }
  }
});
