# Sonar

Agent-first task management for codebases. Designed for AI agents working alongside humans.

A local SQLite-backed CLI that lives in your repo. Agents get structured JSON. Humans get a styled terminal UI. Same commands, same data.

```bash
sonar init
sonar issue create --json -t "Build auth module" -p high -T feature
sonar next --json
```

```json
{"ok": true, "data": {"issues": [...]}, "message": ""}
```

## Install

```bash
go install github.com/ar1o/sonar/cmd/sonar@latest
```

Or build from source:

```bash
make build
```

## How agents use it

Every command supports `--json`. Every response uses the same envelope:

```
{"ok": true,  "data": { ... }, "message": "..."}
{"ok": false, "error": "...",  "code": "NOT_FOUND"}
```

A typical agent loop:

```bash
sonar next --json                          # find unblocked work
sonar issue show SNR-4 --json              # read context
sonar issue move SNR-4 in-progress --json  # claim it
# ... work ...
sonar issue close SNR-4 --json             # done
```

Agents can also decompose work:

```bash
sonar issue create --json -t "Design schema" --parent SNR-1
sonar issue link add SNR-2 depends-on SNR-1 --json
sonar plan --json --root SNR-1
```

`sonar plan` builds a dependency graph and returns a phased execution order. `sonar next` returns only unblocked leaf issues, sorted by priority.

### Add to CLAUDE.md

```
Use `sonar` for issue tracking. Always pass `--json`.
Run `sonar next --json` to find work. Move issues to `in-progress` before starting.
```

## How humans use it

```bash
sonar board                                    # kanban board
sonar issue create -t "Fix login bug" -p high  # create
sonar next                                     # what's ready?
sonar issue show 3                             # details
```

## Commands

```
sonar init                              Initialize .sonar/ database
sonar issue create                      Create an issue
sonar issue list                        List issues (filters, sorting, tree view)
sonar issue show <id>                   Full detail with sub-issues and relations
sonar issue edit <id>                   Edit fields
sonar issue move <id> <status>          Change status
sonar issue close <id>                  Mark done
sonar issue reopen <id>                 Reopen
sonar issue delete <id>                 Delete
sonar issue log <id>                    Activity history
sonar issue comment add <id>            Add comment (-m, stdin, or $EDITOR)
sonar issue comment list <id>           List comments
sonar issue label add <id> <labels>     Add labels
sonar issue label rm <id> <labels>      Remove labels
sonar issue link add <id> <rel> <tid>   Add relation (blocks, depends-on, relates-to, duplicates)
sonar issue link remove <id> <rel> <tid> Remove relation
sonar issue link list <id>              List relations
sonar issue graph <id>                  Dependency graph
sonar issue file add <id> <paths>       Attach files
sonar issue file rm <id> <paths>        Remove attachments
sonar next                              Unblocked work-ready issues
sonar plan                              Phased execution plan from dependency graph
sonar board                             Kanban board
sonar stats                             Database statistics
sonar export                            Export (JSON, CSV, Markdown)
sonar import <file>                     Import from JSON
sonar config                            Show configuration
sonar version                           Version info
```

All commands accept `--json` and `-q` (quiet).

## Concepts

**Statuses:** `backlog` > `todo` > `in-progress` > `review` > `done`

**Priorities:** `critical` | `high` | `medium` | `low` | `none`

**Types:** `bug` | `feature` | `task` | `epic` | `chore`

**IDs:** `SNR-1`, `SNR-42`. Commands accept `SNR-5` or just `5`.

**Relations:** `blocks`, `depends-on`, `relates-to`, `duplicates`. These drive `sonar next` and `sonar plan`.

**Database:** SQLite at `.sonar/issues.db`. Set `SONAR_PATH` to override. Add `.sonar/` to `.gitignore`.

## Development

```bash
make build       # ./bin/sonar
make test        # unit tests
make lint        # staticcheck + go vet
./scripts/qa.sh  # end-to-end test suite
```

## License

MIT
