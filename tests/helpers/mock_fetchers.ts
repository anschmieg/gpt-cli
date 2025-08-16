export function mockFetcher(bodyObj: unknown) {
  return (_input: RequestInfo, _init?: RequestInit) => {
    const res = new Response(JSON.stringify(bodyObj), {
      status: 200,
      headers: { "Content-Type": "application/json" },
    });
    return Promise.resolve(res);
  };
}

export function mockFetcherStream(chunks: string[]) {
  return (_input: RequestInfo, _init?: RequestInit) => {
    const encoder = new TextEncoder();
    const encoded = chunks.map((c) =>
      encoder.encode(
        `data: ${JSON.stringify({ choices: [{ delta: { content: c } }] })}\n\n`,
      )
    );
    const rs = new ReadableStream<Uint8Array>({
      start(ctrl) {
        for (const c of encoded) ctrl.enqueue(c);
        ctrl.enqueue(encoder.encode("data: [DONE]\n\n"));
        ctrl.close();
      },
    });
    const res = new Response(rs, {
      status: 200,
      headers: { "Content-Type": "text/event-stream" },
    });
    return Promise.resolve(res);
  };
}
