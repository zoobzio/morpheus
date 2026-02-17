# Sumatra

A Go application template built on the zoobzio stack.

## The Stack

| Package | Purpose |
|---------|---------|
| `sum` | Service registry, dependency injection, boundaries |
| `rocco` | HTTP handlers, OpenAPI generation, SSE streaming |
| `grub` | Storage abstraction (Database, Bucket, Store, Index) |
| `soy` | Type-safe SQL query builder |
| `pipz` | Composable pipeline workflows |
| `flux` | Hot-reload runtime configuration (capacitors) |
| `cereal` | Field-level encryption, hashing, masking |
| `capitan` | Events and observability signals |
| `check` | Request validation |

## The Crew

This project uses character-driven agents from 3-2-1 Penguins. Each has a specific role and clear boundaries.

### Chain of Command

```
User Request
     │
     ├─→ Unclear requirements, multiple entities, needs planning
     │        ↓
     │     Zidgel (Captain)
     │        │
     │        ├─→ Delegates building/modification to Midgel
     │        ├─→ Delegates testing to Kevin
     │        └─→ Delegates architecture to Fidgel
     │
     ├─→ Build or modify entities
     │        ↓
     │     Midgel (First Mate)
     │
     ├─→ Write tests (unit, integration, benchmarks)
     │        ↓
     │     Kevin (Engineer)
     │
     └─→ Pipeline architecture, documentation, analysis
              ↓
           Fidgel (Science Officer)
```

### When to Invoke Each Agent

**Zidgel** — `@zidgel`
- User requirements are vague or incomplete
- Multi-entity API where relationships need mapping
- User asks "I need a system that..." without specifics
- Planning is required before execution

**Midgel** — `@midgel`
- Clear entity requirements (fields, endpoints defined or easily inferred)
- New entity construction (single or multiple)
- Modifications to existing entities (new fields, new endpoints, new features)
- Any building or extending work

**Kevin** — `@kevin`
- Writing tests for entities
- Unit tests, integration tests, benchmarks
- Test infrastructure (fixtures, mocks, helpers)
- Verifying code works correctly

**Fidgel** — `@fidgel`
- Pipeline design and implementation
- Complex data flow architecture
- Technical documentation
- System analysis and comprehension
- "How does the ingest pipeline work?"
- "Design a processing pipeline for..."

### Direct Invocation Rules

1. **If requirements are clear** — invoke the appropriate agent directly
2. **If requirements are unclear** — invoke Zidgel to extract and plan
3. **Never skip the spec** — all agents produce specs before writing code
4. **Respect the hierarchy** — Zidgel delegates, others execute their domain

## API Surfaces

This project supports multiple API surfaces — separate binaries serving different consumers with different needs.

### Surface Types

| Surface | Binary | Consumer | Characteristics |
|---------|--------|----------|-----------------|
| `api` | `cmd/app/` | Customers | User-scoped, masked data, conservative exposure |
| `admin` | `cmd/admin/` | Internal team | System-wide, full visibility, bulk operations |

### Layer Organization

**Shared layers** (used by all surfaces):
- `models/` — Domain models
- `stores/` — Data access implementations (same store satisfies multiple contracts)
- `migrations/` — Database schema
- `events/` — Domain events
- `config/` — Configuration

**Surface-specific layers** (each surface has its own):
- `{surface}/contracts/` — Interface definitions
- `{surface}/wire/` — Request/response types (different masking per surface)
- `{surface}/handlers/` — HTTP handlers
- `{surface}/transformers/` — Model ↔ wire mapping

### When Working on a Surface

Before creating surface-specific artifacts (handlers, contracts, wire, transformers):

1. **Determine the surface** — Which API is this for?
2. **If unclear, ask** — "Which API surface: public (api/) or admin (admin/)?"
3. **Apply the correct path** — `api/handlers/` vs `admin/handlers/`

Note: Stores are shared. The same store implementation satisfies both `api/contracts.Users` and `admin/contracts.Users`.

### Surface Differences

| Aspect | Public (api/) | Admin (admin/) |
|--------|---------------|----------------|
| Auth | Customer identity | Admin/internal identity |
| Scope | User's own data | System-wide access |
| Operations | Standard CRUD | Bulk ops, impersonation, audit |
| Data exposure | Masked, minimal | Full visibility |

## Skills

Skills live in `.claude/skills/` and define the patterns for each artifact type. They are:

- **Reference material for agents** — agents read skills to understand conventions
- **Directly invocable** — user can call `/add-model` directly for single-layer work
- **Not typically chained by users** — agents handle multi-skill orchestration

### When Users Invoke Skills Directly

| Scenario | Use |
|----------|-----|
| Adding just a migration | `/add-migration` |
| Adding just a config type | `/add-config` |
| User knows exactly what single layer they need | Direct skill |
| User wants full entity/feature | Invoke agent instead |

## Project Structure

```
cmd/
├── app/              # Public API binary
└── admin/            # Admin API binary

# Shared layers
config/               # Static configuration types
models/               # Domain models
stores/               # Data access (shared, satisfies multiple contracts)
events/               # Domain events and signals
migrations/           # Database migrations (goose)
internal/             # Internal packages
testing/              # Test infrastructure, mocks, fixtures

# Public API surface
api/
├── contracts/        # Public interface definitions
├── wire/             # Public request/response types (masked)
├── handlers/         # Public HTTP handlers
└── transformers/     # Public model ↔ wire mapping

# Admin API surface
admin/
├── contracts/        # Admin interface definitions
├── wire/             # Admin request/response types (unmasked)
├── handlers/         # Admin HTTP handlers
└── transformers/     # Admin model ↔ wire mapping
```

## Conventions

### Naming

| Layer | File | Type | Example |
|-------|------|------|---------|
| Model | `models/user.go` | `User` (singular) | `type User struct` |
| Store | `stores/users.go` | `Users` (plural struct) | `type Users struct` |
| Contract | `{surface}/contracts/users.go` | `Users` (plural interface) | `type Users interface` |
| Wire | `{surface}/wire/users.go` | Singular + suffix | `UserResponse`, `AdminUserResponse` |
| Handler | `{surface}/handlers/users.go` | Verb+Singular | `var GetUser`, `var CreateUser` |

### Registration Points

After creating artifacts, wire them appropriately:

**Shared:**
- `stores/stores.go` — aggregate factory (all stores)
- `models/boundary.go` — model boundaries

**Surface-specific (replace `{surface}` with `api` or `admin`):**
- `{surface}/handlers/handlers.go` — `All()` function
- `{surface}/handlers/errors.go` — domain errors
- `{surface}/wire/boundary.go` — wire boundaries (masking for public API)

### Testing

- 1:1 relationship: `user.go` → `user_test.go`
- Helpers in `testing/` call `t.Helper()`
- Mocks use function-field pattern
- Fixtures return test data with sensible defaults

## Response Style

When working in this codebase:
- Follow the agent's character voice if one is invoked
- Produce specs before code for any multi-file work
- Respect layer boundaries — handlers don't import stores directly
- Keep changes minimal — don't over-engineer
