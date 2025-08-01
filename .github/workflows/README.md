# GitHub Actions Workflows

This directory contains GitHub Actions workflows for the SQL Schema Lint MCP Server.

## Workflow

### PR Schema Lint (`pr-schema-lint.yml`)
This workflow connects to an external MCP server to lint SQL schemas using Python. It automatically runs when a Pull Request is opened, synchronized, or reopened that contains changes to SQL files.

## Setup

### Required Secrets

For MCP Server Integration workflows:
- `MCP_SERVER_URL` - The URL of your deployed MCP server (e.g., `https://your-mcp-server.com`)
- `MCP_API_KEY` - (Optional) API key for authenticating with your MCP server

To add secrets:
1. Go to your repository's Settings → Secrets and variables → Actions
2. Click "New repository secret"
3. Add the required secrets

### Customizing Rules
The workflows use the rules defined in `examples/rules/sql-schema-rules.md`. You can:
1. Modify this file to change the default rules
2. Update the workflow to use a different rules file
3. Create PR-specific rules by checking the changed files

## Usage

Once set up, the workflows will automatically run when:
- A Pull Request is opened, synchronized, or reopened
- The PR contains changes to `.sql` or `.ddl` files

The workflow will:
1. Build the MCP server
2. Identify changed SQL files
3. Lint each file against the defined rules
4. Report any issues found
5. Fail the check if errors are found (warnings don't fail the check)

## Example PR Comment
When issues are found, the workflow can automatically comment on the PR with the results.