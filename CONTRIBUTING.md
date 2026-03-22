# Contributing to ByteRoom

Thank you for your interest in contributing! This document explains how to get
started and what we expect from contributors.

---

## Development Workflow

1. Fork the repository and create your branch from `main`:
   ```bash
   git checkout -b feat/your-feature
   ```
2. **Write tests first** (TDD) — see the existing test patterns.
3. Implement the feature.
4. Make sure all tests pass:
   ```bash
   make test
   ```
5. Run the linters:
   ```bash
   make lint
   ```
6. Submit a pull request, filling in the template.

---

## Code Standards

### Go (backend)

- Follow [Effective Go](https://go.dev/doc/effective_go).
- Run `golangci-lint run` before committing.
- Keep packages small, depend inward (domain → infra never infra → domain).
- Errors are returned, not panicked.
- All public symbols have doc-comments.

### TypeScript (frontend)

- Functional components + hooks only (no class components).
- Use Zustand for global state; keep local state local.
- Follow the ESLint config — run `npm run lint` before committing.
- Prefer named exports over default exports for components.

---

## Commit Messages

```
<type>(<scope>): <short description>

[optional body]

[optional footer]
```

| Type | When to use |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `refactor` | No behaviour change |
| `test` | Tests only |
| `docs` | Documentation only |
| `chore` | Build, CI, deps |

**Examples:**
```
feat(api): add message pagination endpoint
fix(ws): reconnect after server-side close
docs(readme): update quick-start instructions
test(hub): add benchmark for 500 concurrent clients
```

---

## Pull Request Checklist

Before submitting, verify:

- [ ] Tests added or updated for your change
- [ ] All tests pass (`make test`)
- [ ] No linting errors (`make lint`)
- [ ] Documentation updated if needed
- [ ] No secrets, keys, or PII committed
- [ ] PR description explains the *why*, not just the *what*

---

## Reporting Issues

Please use GitHub Issues with the appropriate label:

- `bug` — something is broken
- `enhancement` — a new feature request
- `question` — usage or design questions
- `docs` — documentation improvements

Include steps to reproduce and the expected vs. actual behaviour.
