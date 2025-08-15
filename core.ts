import { callProvider } from "./providers/openai.ts";
import { log } from "./utils/log.ts";
import { renderMarkdown } from "./utils/markdown.ts";

export async function runCore(config: any) {
  if (config.verbose) log("Config:", config);
  let response;
  try {
    response = await callProvider(config);
  } catch (err) {
    log("Provider error:", err);
    console.error("Error:", err.message || err);
    Deno.exit(1);
  }
  if (config.verbose) log("Raw response:", response);
  // Output as markdown or plaintext
  if (response.markdown) {
    console.log(renderMarkdown(response.markdown));
  } else {
    console.log(response.text);
  }
}
