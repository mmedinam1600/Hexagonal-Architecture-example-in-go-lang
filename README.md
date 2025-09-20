# HexBank — A Practical Guide to Hexagonal Architecture in Go

> **Goal:** Teach Hexagonal Architecture (Ports & Adapters) with SOLID and Clean Code in a realistic, hands-on Go project.  
> **Scope:** All external dependencies (database, message queue, STP payments) are **simulated** to keep the focus on architecture and design rather than infrastructure.

---

## Why this project exists

Hexagonal Architecture (a.k.a. Ports & Adapters) helps you build software where the **domain and use cases** are independent from **frameworks and infrastructure**. This repository is designed to be a **readable, didactic template** you can clone, run, and extend.

You will learn how to:

- Model a core **banking domain** (Accounts, CLABE) with **business invariants**.
- Implement **use cases** (open account, deposit, transfer) that orchestrate domain rules.
- Define **ports (interfaces)** for outbound dependencies (DB, STP, event bus).
- Implement **adapters** to satisfy those ports (in-memory DB, fake STP with backoff & jitter, local event bus).
- Keep controllers (HTTP handlers) **thin** and focused on I/O and error mapping.
- Apply **SOLID** and **Clean Code** consistently across the codebase.
- Write **tests** that do not require infrastructure to run.

> ⚠️ **Not production-ready.** This repo intentionally uses an in-memory store and a simulated STP service to focus on architecture and teaching. See **“Production considerations”** below.

---

## Features

- **Hexagonal (Ports & Adapters)** layout with clear, one-way dependencies.
- **Domain-first** design: Entities and Value Objects enforce invariants.
- **Use cases** that orchestrate domain behavior (no technical details inside).
- **Ports (interfaces)** for outbound dependencies:
  - `AccountReader`, `AccountWriter` (DB access)
  - `PaymentGateway` (STP)
  - `EventPublisher` (message bus)
- **Adapters**:
  - In-memory repository (thread-safe) to simulate a database.
  - Fake STP client with **exponential backoff + full jitter** retries.
  - Local event bus that logs published events.
- **HTTP API** using Go stdlib (`net/http`), no external frameworks.
- **Shared helpers** for IDs and HTTP JSON responses.
- **Platform helpers** for logging and backoff.
- **Descriptive naming** (no cryptic abbreviations) to ease learning.

---

## Architecture at a glance

```
         ┌───────────────────────────┐
         │       HTTP Adapter        │  (adapters-in: translate HTTP ⇄ DTOs)
         └─────────────┬─────────────┘
                       │ calls
             ┌─────────▼─────────┐
             │    Application     │  (use cases + ports)
             │  - Use Cases       │  open account / deposit / transfer
             │  - Ports           │  AccountReader/Writer, PaymentGateway, EventPublisher
             └─────────┬─────────┘
                       │ depends on interfaces
             ┌─────────▼─────────┐
             │      Domain        │  (entities + value objects + domain errors)
             │  Account, CLABE    │  invariants enforced here
             └────────────────────┘
                       ▲
                       │ ports implemented by adapters-out
  ┌────────────────────┼───────────────────────┐
  │                    │                       │
┌─┴─────────────┐  ┌───┴──────────┐     ┌─────┴─────┐
│  Memory Repo  │  │  Fake STP     │     │ Event Bus │   (adapters-out)
│ (DB simulate) │  │ (backoff)     │     │ (local)   │
└───────────────┘  └───────────────┘     └───────────┘
```

**Dependency rule:** `adapters` depend on `core` (Application + Domain). The core depends **only** on **ports** (interfaces) and **domain** types, never on infrastructure.

---

## Directory structure

```
hexbank/
├─ cmd/bankapp/                         # Entry point (wiring / DI)
│  └─ main.go
└─ internal/
   ├─ core/
   │  ├─ domain/                        # Business rules (pure)
   │  │  ├─ account.go
   │  │  ├─ valueobjects.go             # CLABE as Value Object
   │  │  └─ errors.go
   │  └─ application/                   # Use cases + ports
   │     ├─ ports/
   │     │  └─ ports.go                 # AccountReader/Writer, PaymentGateway, EventPublisher
   │     └─ usecase/
   │        ├─ open_account.go
   │        ├─ deposit_money.go
   │        └─ transfer_money.go
   ├─ adapters/
   │  ├─ in/http/
   │  │  └─ api.go                      # Thin HTTP handlers (stdlib net/http)
   │  └─ out/
   │     ├─ memory/
   │     │  └─ repository.go            # Thread-safe in-memory repo
   │     ├─ stp/
   │     │  └─ fake_stp_client.go       # Fake STP with backoff + jitter
   │     └─ eventbus/
   │        └─ local_event_bus.go       # Logs published events
   ├─ platform/
   │  ├─ backoff/
   │  │  └─ exponential_full_jitter.go  # Backoff policy
   │  └─ logging/
   │     └─ standard_logger.go          # Minimal logger interface + impl
   └─ shared/
      ├─ httpx/
      │  └─ json.go                     # JSON helpers (WriteJSON, WriteError)
      └─ id/
         └─ id.go                       # ID helper (12-byte random hex)
```

---

## Endpoints

All endpoints accept and return JSON.

### Health
```
GET /health  → 200 OK
{ "status": "ok" }
```

### Create account
```
POST /accounts
Content-Type: application/json

{
  "holder_name": "Alice",
  "clabe": "032180000118359719"
}
```
**Response** `201 Created`:
```json
{
  "id": "9c44d0d8f0f340f564b7f1c2",
  "holder_name": "Alice",
  "clabe": "032180000118359719",
  "balance_cents": 0
}
```

### Get account
```
GET /accounts/{id}
```
**Response** `200 OK`:
```json
{
  "id": "9c44d0d8f0f340f564b7f1c2",
  "holder_name": "Alice",
  "clabe": "032180000118359719",
  "balance_cents": 15000
}
```

### Deposit
```
POST /accounts/{id}/deposit
Content-Type: application/json

{ "cents": 15000 }
```
**Response** `200 OK`:
```json
{ "id": "…", "balance_cents": 15000 }
```

### Transfer
```
POST /transfers
Content-Type: application/json

{ "from_id": "…", "to_id": "…", "cents": 5000 }
```
**Response** `202 Accepted`:
```json
{ "from_balance_cents": 10000, "to_balance_cents": 5000 }
```

---

## Error handling

**Domain errors** are mapped by the HTTP adapter into HTTP codes:

- `ErrInvalidAmount` → **400 Bad Request**
- `ErrInvalidCLABE`, `ErrEmptyHolder` → **422 Unprocessable Entity**
- `ErrInsufficientFund` → **422 Unprocessable Entity**
- Any unexpected error → **500 Internal Server Error**

Example:
```json
{ "error": "insufficient funds" }
```

---

## How to run

**Prerequisites:** Go 1.20+ (repo uses stdlib only).

```bash
go version                 # verify Go installation
go mod tidy                # sync dependencies (none external, still safe to run)
go run ./cmd/bankapp       # start HTTP API on :8080
```

**Sample usage:**

```bash
# Create account
curl -sS -X POST http://localhost:8080/accounts   -H "Content-Type: application/json"   -d '{"holder_name":"Alice","clabe":"032180000118359719"}'

# Get account
curl -sS http://localhost:8080/accounts/<ID>

# Deposit
curl -sS -X POST http://localhost:8080/accounts/<ID>/deposit   -H "Content-Type: application/json"   -d '{"cents":15000}'

# Create a second account
curl -sS -X POST http://localhost:8080/accounts   -H "Content-Type: application/json"   -d '{"holder_name":"Bob","clabe":"032180000118359700"}'

# Transfer
curl -sS -X POST http://localhost:8080/transfers   -H "Content-Type: application/json"   -d '{"from_id":"<ALICE_ID>","to_id":"<BOB_ID>","cents":5000}'
```

When transferring, the **Fake STP** might simulate transient failures. Retries use **exponential backoff + full jitter**, and events are logged by the **Local Event Bus** upon success.

---

## How Hexagonal + SOLID + Clean Code are applied

### Hexagonal (Ports & Adapters)
- The **core** (`internal/core`) does not import frameworks or infrastructure.
- **Ports** live in `application/ports` and define what the use cases need.
- **Adapters** in `internal/adapters` satisfy the ports for a particular technology.
- **Entry point** (`cmd/bankapp`) wires everything together.

### SOLID
- **S — Single Responsibility:** Each file/class has one job: entities protect invariants, each use case orchestrates one goal, adapters talk to one technology.
- **O — Open/Closed:** Add a new database or real STP by creating a new adapter. The core stays closed for modification but open for extension via ports.
- **L — Liskov Substitution:** Use cases depend on interfaces; any compliant implementation (fake/real) can be substituted.
- **I — Interface Segregation:** `AccountReader` and `AccountWriter` are separate; small, focused ports.
- **D — Dependency Inversion:** Core depends on **abstractions** (ports); concrete adapters depend on the core, not the other way around.

### Clean Code
- **Descriptive names** (no cryptic abbreviations).
- **Thin controllers** (translate HTTP ⇄ DTO + error mapping).
- **Pure domain** (no I/O in entities/VOs).
- **Explicit invariants** and **domain-level errors**.
- **Helpers** isolated in `shared/` and cross-cutting concerns in `platform/`.

---

## Backoff strategy (exponential full jitter)

Implemented in `internal/platform/backoff/exponential_full_jitter.go` and used by the fake STP adapter.

- **Parameters** (in the adapter):
  - `maxRetries = 4`
  - `baseDelay = 200ms`, `multiplier = 2.0`
  - `maxDelay = 3s`
- **Why full jitter?** Reduces thundering herd and provides better tail latency than fixed or equal jitter.

> To tune behavior, change those constants in `adapters/out/stp/fake_stp_client.go`.

---

## Testing

Unit test example (domain invariants):
```bash
go test ./internal/core/domain -v
```

Run **all** tests:
```bash
go test ./... -v
```

**Guidelines:**
- **Domain** tests do not require any infrastructure.
- **Use case** tests should use **fakes** for ports.
- **Adapter** tests may be integration-style (e.g., check repository behavior).

---

## How to extend

### Add a new use case (example: WithdrawMoney)
1. **Domain**: If you need new invariants, add them to `Account`.
2. **Application**: Create `withdraw_money.go` in `usecase` with input/output DTOs and the orchestration.
3. **Ports**: Reuse `AccountReader/Writer`. Add new ports only when you truly need a new dependency.
4. **Adapters-in**: Add an HTTP handler/route in `adapters/in/http/api.go`.
5. **Wiring**: Update `cmd/bankapp/main.go` if you add new dependencies.
6. **Tests**: Add unit tests for the domain and the new use case with fakes.

### Swap the database
- Create a new package under `adapters/out` (e.g., `postgres/`).
- Implement `AccountReader` and/or `AccountWriter`.
- Wire it in `main.go`. The core code stays the same.

### Use a real STP client
- Create a new adapter that implements `PaymentGateway` using `net/http`.
- Keep retry/backoff policy in the adapter (infra concern).
- Consider **idempotency keys** for production-grade transfers.

---

## Production considerations (out of scope here)

- **Persistence:** Use a real database; add migrations and a `UnitOfWork` pattern if needed.
- **Idempotency:** Required for transfer requests (e.g., header `Idempotency-Key`).
- **Observability:** Metrics (latency, retries), tracing, structured logs.
- **Security:** Authentication/authorization, input validation, secrets management.
- **Error model:** Dedicated error types and mapping strategy.
- **Configuration:** Environment variables / config files.
- **CI/CD:** Lint (`go vet`, staticcheck), tests, and reproducible builds.
- **Contracts:** OpenAPI/Swagger for the HTTP API.

---

## Commands you will use often

- `go mod tidy` — Synchronize `go.mod`/`go.sum` with imports; add missing and remove unused deps.
- `go build ./...` — Compile the entire module (good smoke test).
- `go run ./cmd/bankapp` — Run the HTTP API locally.
- `go test ./...` — Run all tests.

---

## License

This project is distributed under the **GNU General Public License (GPL)**, as indicated by the `LICENSE` file in the repository root.

---

## Contributing

PRs are welcome! Please follow these guidelines:

- Use **descriptive names** and keep functions small & cohesive.
- Keep adapters thin; put **business rules** in the **domain**.
- Write tests for **domain** and **use cases**; avoid hitting real networks.
- Maintain the **dependency direction** (core ← adapters).

---

## FAQ

**Why in-memory DB?**  
To keep the architecture front-and-center. Swap for Postgres/Mongo by writing a new adapter that implements the ports.

**Why stdlib net/http?**  
To avoid framework noise. You can add your favorite router later without touching the core.

**What is CLABE?**  
A Value Object representing a Mexican banking standard. Here we only enforce “18 digits” to illustrate invariants.

**What about money types?**  
We store **cents** as `int64` to avoid floating-point issues. In a real system, consider a dedicated `Money` value object (currency + amount).

---

Happy hacking! If this project helps you, consider sharing the repo and teaching others what you learned about Hexagonal Architecture in Go.
