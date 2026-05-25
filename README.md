# devcli 🚀

`devcli` is a Go command-line tool designed to supercharge developer rest API testing by defining dynamic, templated API targets. It registers custom rest subcommands from local repository definitions, auto-generates CLI flags from placeholders, supports multi-profile switches, and pretty-prints syntax-highlighted response payloads.

## Features

- ⚙️ **Unified Configuration**: Both profiles and subcommand endpoints are stored together in a single config file.
- 🔑 **Multi-file Fallback**: Looks for a local `.devcli.yaml` first (great for sharing API definitions in a repository), and falls back to a global config (`~/.config/devcli/config.yaml`) if none exists.
- 📝 **Template Interpolation**: Paths, query params, request headers, and JSON request bodies auto-generate kebab-case flags from placeholders (`{user_id}` -> `--user-id`).
- 🌈 **Syntax-Highlighted Output**: JSON responses are automatically pretty-printed and syntax-colored.
- 🧪 **Smart JSON Cleaning**: Unused optional parameters in body templates are automatically removed so JSON payloads remain syntactically valid.

---

## Installation

Compile the binary from the root directory:

```bash
go build -o devcli
```

*(Optional)* Move the compiled binary into your `$PATH` (e.g. `/usr/local/bin`) for global access.

---

## Configuration

`devcli` loads a single unified configuration file containing both `profiles` and `commands`:

1. **Local Configuration** (`.devcli.yaml` in your project root): Checked first. Ideal if you want to write and share command templates (and project-specific targets) with your team via git.
2. **Global Configuration** (`~/.config/devcli/config.yaml`): Checked if no local configuration is found. Useful for personal/global profiles.

When configuring profiles via commands (e.g. `devcli profile set`), the values are saved directly back to the active configuration file (`.devcli.yaml` or global `config.yaml`).

### Configuration Schema

Here is an example of a unified `.devcli.yaml` configuration containing both target profiles and commands:

```yaml
current_profile: development
profiles:
  development:
    url: "http://localhost:8080"
    token: "dev-token-xyz"
    headers:
      X-App-ID: "dev-app"

commands:
  get-user:
    path: "/users/{id}"
    method: "GET"
    description: "Fetch a user profile by ID"
    query:
      fields: "{fields}"
      v: "1"
    headers:
      X-Client-Header: "devcli"

  create-user:
    path: "/users"
    method: "POST"
    description: "Create a new user"
    body: |
      {
        "name": "{name}",
        "email": "{email}",
        "role": "{role}"
      }
    args:
      name:
        required: true
        description: "The name of the user"
      email:
        required: true
        description: "The email address"
      role:
        default: "developer"
        description: "User authority role"
```

---

## Command Usage Examples

Once configured, your dynamic API subcommands can be called directly:

```bash
# Call a GET request (automatically requires --id because it's in the path)
devcli get-user --id 123 --fields "name,email"

# Use verbose logging to print raw request/response headers
devcli get-user --id 123 --verbose

# Call a POST request (role will default to "developer")
devcli create-user --name "John Doe" --email "john@example.com"

# Override defaults in body parameters
devcli create-user --name "Alice" --email "alice@example.com" --role "admin"
```