export async function callProvider(config: any): Promise<{ text?: string, markdown?: string }> {
  const apiKey = Deno.env.get("OPENAI_API_KEY");
  if (!apiKey) throw new Error("OPENAI_API_KEY not set in environment");
  const url = "https://api.openai.com/v1/chat/completions";
  const body = {
    model: config.model || "gpt-3.5-turbo",
    messages: [
      config.system ? { role: "system", content: config.system } : null,
      { role: "user", content: config.prompt },
    ].filter(Boolean),
    temperature: config.temperature || 1.0,
    max_tokens: 2048,
  };
  const res = await fetch(url, {
    method: "POST",
    headers: {
      "Authorization": `Bearer ${apiKey}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const err = await res.text();
    throw new Error(`OpenAI API error: ${res.status} ${err}`);
  }
  const json = await res.json();
  const text = json.choices?.[0]?.message?.content || "";
  return { text };
}
