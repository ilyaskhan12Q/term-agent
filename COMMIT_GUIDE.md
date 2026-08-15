# Commit Guide

## Conventions

`term-agent` follows Conventional Commits version 1.0.0.

### Commit Format

```text
<type>(<scope>): <short summary>

[optional body]

[optional footer(s)]
```

### Types

- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation updates only
- `style`: Formatting, missing semi-colons, whitespace fixes (`gofmt`)
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `test`: Adding missing tests or correcting existing tests
- `chore`: Infrastructure, dependencies, build tasks

### Rules

1. Do not use emojis in commit messages or pull request titles.
2. Use lowercase for type and scope.
3. Keep the first line under 72 characters.
4. Reference issue or PR numbers in the footer when applicable.
