<p align="center">
  <img src="docs/assets/logo.png" alt="mossy" width="200" />
</p>

<h1 align="center">🌿 mossy</h1>

<p align="center">
  A terminal dashboard for your GitHub repositories, built with <a href="https://github.com/charmbracelet/bubbletea">Bubble Tea</a>.
</p>

---

## Features

- **Tab bar** — Switch between registered GitHub repositories with `h`/`l`
- **Repo picker** — Browse your filesystem and add git repos with `a`
- **Git detection** — Only directories with `.git` can be added

## Install

```sh
go install github.com/marcellolins/mossy@latest
```

## Usage

```sh
mossy
```

### Key Bindings

| Key | Action |
|---|---|
| `a` | Add a repository |
| `d` | Remove a repository |
| `n` | New worktree |
| `x` | Remove worktree |
| `[` / `]` | Prev / next commit |
| `h` / `l` | Switch tabs |
| `j` / `k` | Navigate lists |
| `r` | Refresh |
| `R` | Toggle auto-refresh |
| `enter` | Select / open directory |
| `esc` | Cancel |
| `?` | Help |
| `q` | Quit |
| `ctrl+c` | Force quit |

## License

[MIT](LICENSE)
