<p align="center">
  <img src="docs/assets/banner.png" alt="Vanitas" width="100%">
</p>

<h3 align="center">Read the true names of your data.</h3>

<p align="center">
  A terminal database client for PostgreSQL, MySQL/MariaDB and SQLite.<br>
  One static binary. Safe by default. Shortcuts always on screen.
</p>

---

> **Status: early development.** Nothing is released yet. This README describes the v1 target;
> features land phase by phase.

## What it is

Vanitas is a TUI database client that brings together three ideas:

- **Pick up and use** (inspired by [sqlit](https://github.com/Maxteabag/sqlit)): run `vanitas`,
  pick a connection, start working. Every screen shows its main shortcuts in the footer, and `?`
  opens the help for the current screen.
- **Edit data in the grid** (inspired by [lazysql](https://github.com/jorgerojas26/lazysql)):
  change cells, add and delete rows; changes stay pending until you save, with a preview of the
  generated SQL, applied in a single transaction.
- **An IDE-like SQL editor** (inspired by [harlequin](https://github.com/tconbeer/harlequin)):
  syntax highlighting, Vim mode, autocomplete, buffers that persist between sessions, run the
  statement under the cursor or the whole buffer.

## Planned features (v1)

- Connection manager: saved connections, URLs (`vanitas postgres://...`, `vanitas ./app.db`),
  passwords in the OS keyring or a `password_command`, SSH tunnels.
- Schema explorer with lazy loading, fuzzy filter and quick search (`Ctrl+P`).
- Results grid that stays fluid with millions of rows: pagination, streaming, zebra rows,
  distinct `NULL`, server-side `WHERE` filter and sorting.
- Foreign-key display values: `team_id` shows `42 → Platform Team`.
- Table structure: columns, constraints, foreign keys, indexes, DDL.
- Copy as CSV, JSON, Markdown or `INSERT`; streaming CSV export.
- Read-only and production connections, query timeouts, confirmation for dangerous statements.
- Leader menu (`Space`), customisable keymap, 8 themes, mouse support, `NO_COLOR`.
- Linux, macOS and Windows; amd64 and arm64; no CGO, no runtime dependencies.

Out of scope for v1: other databases (SQL Server, DuckDB, Oracle...), Docker discovery, a
headless CLI, Parquet/JSON export, plugins.

## Why "Vanitas"

The name comes from *The Case Study of Vanitas* (*Vanitas no Karte*). In that story, a vampire's
curse is a corruption of its **true name**, and the Book of Vanitas can read that name and
restore it.

Databases have true names too: their real structure, their real data, their real
relationships. Over time they get corrupted by inconsistent rows, orphaned foreign keys and
forgotten queries. Vanitas reads a database the way the Book reads a vampire: it shows what is
really there and lets you restore it safely.

*Every database has a true name. Read. Understand. Restore.*

The theme lives in the look and in a few words of interface text, never in your way. Commands,
shortcuts and documentation stay plain and standard.

| In the story | In Vanitas |
|---|---|
| The Grimoire | The query editor, with buffers that persist between sessions |
| The true name | The table structure view: columns, indexes, foreign keys |
| Reading | Browsing and filtering data in the grid |
| Restoring | Editing in the grid: changes stay pending until `Ctrl+S`, with an SQL preview |
| A curse | A query or constraint error, shown prominently |
| The Paladin | Read-only mode and production connections, which protect the database from you |
| The Archiviste | Query history and data export |
| The blue moon | The default colour theme: midnight blue with gold accents |

## Installation

Not available yet. Releases will ship as single binaries for Linux, macOS and Windows (built
with GoReleaser). Until then, from source, with [mise](https://mise.jdx.dev):

```sh
git clone https://github.com/allanmedeiros71/vanitas.git
cd vanitas
mise install
make build
./vanitas version
```

## Usage (planned)

```sh
vanitas                              # open the connection picker
vanitas postgres://user@host:5432/db # temporary connection from a URL
vanitas ./notes.db                   # SQLite file
vanitas -c local-pg                  # open a saved connection
vanitas connections list|add|remove
vanitas version
```

Configuration lives in `$XDG_CONFIG_HOME/vanitas` (`~/.config/vanitas`; macOS
`~/Library/Application Support/vanitas`; Windows `%APPDATA%\vanitas`), or in
`VANITAS_CONFIG_DIR`. Passwords are never stored in plain text.

## Privacy

Vanitas has no telemetry and sends nothing anywhere except to the databases you connect to.
Passwords, URLs with passwords and row values are never written to logs. Query history is kept
locally and can be deleted.

## License

[MIT](LICENSE)
