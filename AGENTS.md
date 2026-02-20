## Build

Always run `go build -o mossy .` at the end of each prompt to ensure the binary is up to date.

## Layout: Footer Bar Positioning

The TUI layout in `ui.go` composes three sections: `top + "\n" + content + "\n" + foot`.

Key rules to avoid the footer disappearing:

1. **Content must be exactly `mid` lines with NO trailing newline.** The worktree list View uses `strings.Split`/`strings.Join` to produce exactly `height` lines. The `"\n"` separators in the composition join the sections without adding extra lines.
2. **Content must be both padded AND truncated.** If content has fewer lines than `mid`, pad with empty strings. If it has more, truncate to `mid`. Never allow content to exceed the allocated height or the footer gets pushed off-screen.
3. **The height budget is `mid = Height - topHeight - footHeight`.** No extra subtraction needed — the two `"\n"` separators between top/content and content/foot just join the strings, they don't add extra visual lines (since none of the sections end with a trailing newline).
