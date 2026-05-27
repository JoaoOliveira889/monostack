# Rules for Monostack (AWS TUI)

This document establishes the architectural guidelines, design principles, and coding standards for **Monostack**, ensuring strict adherence to the Clean Architecture patterns established in the project.

---

## 1. Architectural Integrity (Clean Architecture)

We strictly follow the Ports & Adapters (Hexagonal / Clean) Architecture pattern. The dependency flow must only point **inward**:
`Adapters` ──> `Usecase` ──> `Domain`

### 1.1 `internal/domain` (Core Business Logic)
- **Rules**:
  - **No external dependencies** (except Go standard library and possibly generic metadata packages).
  - Must **never** import `internal/usecase`, `internal/adapters`, or `cmd`.
  - Defines core business structs (e.g., `Bucket`, `Object`, `Queue`, `Message`, `Topic`, `Config`).
  - Defines **Port Interfaces** (e.g., `S3Manager`, `SQSManager`, `SNSManager`, `ConfigStore`).

### 1.2 `internal/usecase` (Orchestration & Business Rules)
- **Rules**:
  - Orchestrates domain entities and executes business logic.
  - Can only import `internal/domain` and `internal/pkg`.
  - Must **never** import `internal/adapters` or `cmd`.
  - Coordinates calls to Port Interfaces without knowing the underlying implementation details.

### 1.3 `internal/adapters` (External Frameworks & Services)
- **Rules**:
  - Implements the interfaces defined in the domain layer.
  - May import `internal/domain`, `internal/usecase`, and `internal/pkg`.
  - Consists of:
    - `aws/`: Concrete implementations of `S3Manager`, `SQSManager`, and `SNSManager` using AWS SDK for Go V2, fully supporting dynamic endpoint routing (for `mini-stack` / LocalStack).
    - `tui/`: Bubble Tea adapters including Model, Update, View, and styled components.
    - `config/`: Concrete file-based configurations storage/loading.

### 1.4 `internal/pkg` (Common/Shared Utilities)
- **Rules**:
  - Contains generic utilities (e.g., formatters, standard tools).
  - Must be fully independent of the rest of the application layers.

---

## 2. Bubble Tea (TUI) Architecture & Concurrency

To ensure the terminal interface is fast, completely responsive, and free of blocking, follow these rules:

- **Non-Blocking / Asynchronous Operations**: 
  - All external AWS API calls (listing buckets, fetching queue counts) **must** run asynchronously using Bubble Tea commands (`tea.Cmd`).
  - Never perform network or disk IO inside `Update()` directly. Always return a `tea.Cmd` that performs the operation and returns a success/error message.
- **State Separation**:
  - Separate views into individual files inside `internal/adapters/tui/` (e.g., `view_panels.go`, `view_components.go`).
  - Maintain a clean elm-architecture pattern where `Update` processes standard structural messages and triggers view changes.
- **Premium Aesthetics (Lip Gloss & Styles)**:
  - Custom color palettes only. No generic blue or red terminal colors. Use carefully designed HSL or hex color mappings.
  - Implement dynamic window resizing handled inside `tea.WindowSizeMsg`.

---

## 3. Go & AWS SDK Best Practices

- **SDK v2 Usage**: Always use AWS SDK for Go v2 packages (`github.com/aws/aws-sdk-go-v2/...`).
- **Endpoint Flexibility**: Ensure all clients can resolve dynamic endpoints through `aws.EndpointResolverWithOptionsFunc` to enable smooth mapping to `mini-stack`, LocalStack, or mock servers.
- **Context Handling**: Pass explicit `context.Context` from the adapters down to the SDK calls for query cancellation and timing control.
- **Error Propagation**: Return typed/wrapped domain errors from the adapters, so the usecase/TUI layer can gracefully handle and display styled notification/error banners to the user.
