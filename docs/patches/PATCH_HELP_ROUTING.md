
# Help & Man Routing Patch

Adds support for:

- restless help
- restless help <topic>
- restless --man
- restless <command> --help

Routes to entry.Help() and entry.Man() (must exist).

This patch only modifies cmd/restless/main.go.
