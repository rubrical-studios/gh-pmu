# Windows Shell Safety for Claude Code
**Version:** v0.42.0
**Source:** Reference/Windows-Shell-Safety.md
**Source:** Reference/Windows-Shell-Safety.md
---
**MUST READ:** Auto-loaded on Windows at session startup.
## Shell Environment
Claude Code uses Git Bash on Windows. Most Unix commands work, but these patterns fail.
## Golden Rules
**Always use Unix-style commands and patterns.**
### Command Translations
| Windows CMD | Unix/Bash |
|-------------|-----------|
| `date /t` | `date "+%Y-%m-%d"` |
| `dir` | `ls` |
| `copy` | `cp` |
| `move` | `mv` |
| `del` | `rm` |
| `type` | `cat` (or Read tool) |
| `%VARIABLE%` | `$VARIABLE` |
| `set VAR=value` | `export VAR=value` |
| `C:\path\to\file` | `C:/path/to/file` |
| `taskkill /F /IM` | `powershell -Command 'Stop-Process ...'` |
### Pattern Safety
| Pattern | Safe? | Use Instead |
|---------|:-----:|-------------|
| `$(command)` anywhere | ❌ | Write tool + temp file |
| `for x in $(cmd)` | ❌ | Glob patterns or temp file |
| Nested `$(...)` | ❌ | Sequential commands |
| Heredoc with backticks | ❌ | Write tool + temp file |
| `--body "..."` inline | ❌ | `-F file.md` or `--body-file` |
| `--flag value` (strings) | ⚠️ | `--flag=value` |
| Backslash paths | ❌ | Forward slashes |
| Absolute paths in args | ❌ | Relative paths (`.tmp-*`) |
| JSON inline | ❌ | `--input` or temp file |
| `rm -rf a/ b/ c/` | ❌ | One path at a time |
| Single quotes | ⚠️ | Prefer double quotes |
| Pipes `\|` | ✅ | — |
| Redirects `>` `>>` | ✅ | — |
| `$VAR` expansion | ✅ | — |
| `--body-stdout` / `--body-stdin` | ✅ | Use `.tmp-{issue#}.md` for edits |
### Parallel Tool Failures
**"Sibling tool call errored" is NOT the real error.**
When Claude Code runs multiple tools in parallel and one fails, all siblings are aborted. Find the ONE tool with the actual error, fix it, retry.
```
● Bash(date /t)                    ← ROOT CAUSE (find this)
  ⎿  Error: date: invalid date '/t'

● Bash(git branch)                 ← Aborted (ignore)
  ⎿  Error: Sibling tool call errored
```
## Heredocs with Backticks
**NEVER use backticks inside heredocs.**
```bash
# BAD - fails on Windows
git commit -m "$(cat <<'EOF'
Fix bug in `calculateTotal` function
EOF
)"

# GOOD - use Write tool to create temp file
git commit -F .tmp-msg.txt
rm .tmp-msg.txt
```
## Command Substitution
**ALL `$(...)` patterns are unreliable:** tool arguments, variable assignments, for loops, nested substitution.
```bash
# BAD - unreliable
gh issue create --body "$(cat README.md)"

# GOOD - use --body-file flag
gh issue create --body-file README.md
```
## Issue/PR Bodies with Backticks (IMPORTANT)
**ALWAYS use temp file approach for issue/PR bodies.** Bodies almost always contain backticks (code blocks, inline code).
```bash
# BAD - backticks cause failures
gh pmu create --body "Fix \`foo\` function"

# GOOD - Write tool + temp file
gh pmu create --title "Bug: ..." -F .tmp-body.md --status backlog
rm .tmp-body.md
```
**Rule:** Never use `--body` with inline content. Always use `-F` with a temp file.
## gh pmu Body Flags
**Prefer `--body-stdout` / `--body-stdin`** for cleaner workflows.
```bash
# Export body, edit, update (use issue-specific temp file name)
gh pmu view 123 --body-stdout > .tmp-123.md
gh pmu edit 123 -F .tmp-123.md && rm .tmp-123.md

# Creating issues with body from file (new issue, no number yet)
gh pmu create --title "Bug: ..." -F .tmp-body.md --status backlog

# Alternative: Body-File Pattern (uses tmp/ directory)
gh pmu view 123 --body-file    # Creates tmp/issue-123.md
gh pmu edit 123 -F tmp/issue-123.md
rm tmp/issue-123.md
```
## Path Handling
**Use forward slashes. Quote paths with spaces.**
```bash
# BAD - backslashes stripped
cd C:\Users\Name\My Projects

# GOOD
cd "C:/Users/Name/My Projects"
cd "$USERPROFILE/My Projects"
```
## Temp File Best Practices
1. **Use relative paths** for temp files (`.tmp-*`) - absolute paths get backslashes stripped
2. **Use Write tool** instead of `cat`, `echo >`, or heredocs
3. **Clean up** immediately after use
4. **Use unique names** when editing multiple issues - include issue number (`.tmp-123.md`)
```bash
# Pattern: Write tool creates file, Bash uses it
gh issue create --body-file .tmp-body.md
rm .tmp-body.md
```
## Quoting Rules
**Prefer double quotes. Escape special characters.**
```bash
# GOOD
echo "$HOME"
echo "The file is \`important\`"
```
## JSON in Commands
**Use temp files for JSON payloads.**
```bash
# GOOD - Write tool creates JSON file
gh api graphql --input .tmp-query.json
rm .tmp-query.json
```
## Flag Values with Spaces
**Flags taking string values can fail when space-separated.**
`--flag value` can be misinterpreted as `--flag` (no value) plus `value` (extra argument).
```bash
# BAD - value interpreted as separate argument
gh pmu view 3 --json status --jq '.status'
# Error: accepts 1 arg(s), received 2

# GOOD - use = to attach value directly
gh pmu view 3 --json=status --jq='.status'
gh pmu view 3 --json=number,title,status
```
**Use `=` syntax for:** single-value string flags like `--json=status`, `--jq='.field'`
**Safe rule:** When in doubt, use `--flag=value`.
## Loops with Command Substitution
**Loops using `$(...)` fail.** File globbing works.
```bash
# BAD - for loop with command substitution
for file in $(find . -name "*.md"); do echo "$file"; done

# GOOD - glob patterns work
for file in *.md; do echo "$file"; done

# GOOD - pre-compute to temp file
git log --oneline > .tmp-list.txt
while read -r line; do echo "$line"; done < .tmp-list.txt
rm .tmp-list.txt

# GOOD - find with -exec
find . -name "*.md" -exec wc -l {} \;
```
## Environment Variables
**Use Unix-style syntax in Git Bash.**
```bash
# BAD - Windows cmd style
echo %USERPROFILE%

# GOOD - Unix style
echo "$USERPROFILE"
export MY_VAR=value
```
## Dangerous rm Patterns
**NEVER use `rm -rf` with multiple paths on Windows Git Bash.**
```bash
# DANGEROUS
rm -rf .vite/ out/ dist/

# SAFE - one at a time
rm -rf .vite
rm -rf out
rm -rf dist
```
## Process Management
**Use PowerShell for killing processes on Windows Git Bash.**
Git Bash mangles Windows command flags (interprets `/F` as a path).
```bash
# BAD - Git Bash interprets /F as path
taskkill /F /IM "electron.exe"

# GOOD - Single quotes pass command verbatim to PowerShell
powershell -Command 'Stop-Process -Name "electron" -Force -ErrorAction SilentlyContinue'
```
**Recommended cleanup sequence:**
```bash
# 1. Kill app processes
powershell -Command 'Stop-Process -Name "YourAppName" -Force -ErrorAction SilentlyContinue'

# 2. Kill stray Electron processes
powershell -Command 'Stop-Process -Name "electron" -Force -ErrorAction SilentlyContinue'

# 3. Wait for file handles to release
sleep 2

# 4. Clean build directories (one at a time!)
rm -rf .vite
rm -rf out
```
**Detection:** "Device or resource busy" errors mean processes are still running.
---
**End of Windows Shell Safety**
