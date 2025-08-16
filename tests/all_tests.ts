// Entrypoint for all tests in the gpt-cli project
// Import all test files so Deno can run them from a single entrypoint

import "./config_test.ts";
import "./core_test.ts";
import "./utils_test.ts";

// CLI tests
import "./cli/cli_args_edge_test.ts";
import "./cli/cli_args_test.ts";
import "./cli/cli_parse_test.ts";

// Providers tests
import "./providers/adapter_utils_test.ts";
import "./providers/api_openai_compatible_extra_test.ts";
import "./providers/openai_chat_test.ts";
import "./providers/openai_local_test.ts";
import "./providers/provider_adapter_copilot_test.ts";
import "./providers/provider_adapter_gemini_test.ts";
import "./providers/provider_adapter_openai_test.ts";
import "./providers/provider_adapter_shape_test.ts";
import "./providers/provider_stream_test.ts";
import "./providers/provider_unit_test.ts";

// Core tests
import "./core/retry_model_test.ts";
import "./core/runCore_adapter_validation_test.ts";
import "./core/runCore_fallback_test.ts";
import "./core/runCore_retry_and_fallback_test.ts";
import "./core/runCore_stream_test.ts";

// Integration tests
import "./integration/provider_integration_test.ts";
import "./integration/provider_test.ts";
import "./integration/streaming_integration_test.ts";

// Helpers (if needed)
// import "./helpers/mock_fetchers.ts";
