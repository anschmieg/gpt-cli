#!/usr/bin/env bash
set -euo pipefail
exec go run ./mock-openai "$@"
