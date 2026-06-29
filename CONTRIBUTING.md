# Contributing

Issues are welcome, pull requests are better.

## Pull requests

- Keep PRs **small and atomic** — a human must be able to verify the change by reading it.
- Write a clear description: what problem it solves and how.
- `golangci-lint run` must pass. If you don't have it locally, CI will catch it.
- Code should be **Go-idiomatic** — the reference is the Go standard library. When in doubt, match the existing style in the repo.

## LLM-generated contributions

PRs produced with the help of an LLM are accepted under the same rules. The contributor
is responsible for the output — review it before submitting. Submitting unreviewed
LLM output wastes reviewer time; repeated offenders will be banned from contributing.

## Security

This project handles authentication and authorization. Prefer **obvious and
straightforward** solutions over clever or complex ones. If a reviewer can't
convince themselves it's correct in one pass, it's probably too subtle.

## MCP for agents


### Frontend

- vue docs: https://vue-mcp.org/guide/getting-started
- shadcn: https://ui.shadcn.com/docs/mcp