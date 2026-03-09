Always respond in Chinese-simplified

# LGH Agent Protocol

This repository is the LGH codebase for Town 2.0.

## Town Role

- `project_id = LGH`
- LGH is the local git and LAN sync infrastructure layer in Town 2.0.
- In the Town stack, LGH is part of the control-center candidate together with ActionD.

## First Read

- [`/Users/fenge1222/joe/town-v2/README.md`](/Users/fenge1222/joe/town-v2/README.md)
- [`/Users/fenge1222/joe/town-v2/docs/townV2-rfc-0001.md`](/Users/fenge1222/joe/town-v2/docs/townV2-rfc-0001.md)
- [`/Users/fenge1222/joe/town-v2/docs/townV2-rfc-0002-multi-window-governance.md`](/Users/fenge1222/joe/town-v2/docs/townV2-rfc-0002-multi-window-governance.md)

## Default Shared Scope

- `resident_id = codex`
- `app_id = codex`
- `user_id = joe`

## Multi-Window Rules

- Treat this window as one project cell unless explicitly acting as the main window.
- Before substantial work, recall shared memory for `LGH`, `control center`, and the current task.
- Use explicit `window_id` and `task_id` in notes or handoff when continuity matters.
- Do not assume another window knows local findings unless they are written to RMS or committed to artifacts.
- Before ending substantial work, write a concise handoff covering changes, state, pending work, and main risk.

## Repository Focus

- Preserve LGH as a stable code-state and LAN-sync layer.
- Prefer explicit remotes and auditable git flows over hidden automation.
- Avoid destructive git operations unless explicitly requested.
