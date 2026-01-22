# Windows Shell Safety for Claude Code
**Version:** v0.29.3
**Source:** Reference/Windows-Shell-Safety.md
---
**MUST READ:** Auto-loaded on Windows at session startup.
## Shell Environment
Claude Code uses Git Bash on Windows. Most Unix commands work, but these patterns fail.
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
# Export body, edit, update (preferred - no tmp/ directory)
gh pmu view 123 --body-stdout > .tmp-body.md
gh pmu edit 123 -F .tmp-body.md && rm .tmp-body.md

# Creating issues with body from file
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
## Quick Reference
| Pattern | Windows Safe? | Alternative |
|---------|:-------------:|-------------|
| `$(command)` anywhere | Unreliable | Write tool + temp file |
| `for x in $(cmd)` loops | No | Glob patterns or temp file |
| Nested `$(...)` | No | Sequential commands |
| Heredoc with backticks | No | Write tool + temp file |
| Single quotes | Mostly | Prefer double quotes |
| Backslash paths | No | Forward slashes |
| Absolute paths in Bash args | No | Relative paths (`.tmp-*`) |
| `--body "..."` multi-line | Unreliable | `--body-file` |
| `--body-stdout` / `--body-stdin` | Yes | - |
| JSON inline | No | `--input` or temp file |
| Pipes `\|` | Yes | - |
| Redirection `>` `>>` | Yes | - |
| `$VAR` expansion | Yes | - |
---
**End of Windows Shell Safety**
