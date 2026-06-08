# Strategic Decisions (SDR)

This file tracks major strategic tradeoffs and architectural decisions made by the User, including *why* certain paths were chosen or rejected.

## SDR-001: Separation of Director and Symphony
* **Status**: Decided
* **Context**: We need a way to track strategic direction without cluttering our actual runtime code, and without coupling orchestration to application-level goals.
* **Decision**: We split these concerns into two conceptual layers. "Director" owns strategic continuity and has no runtime code. "Symphony" owns orchestration mechanics.
* **Consequences**:
  * We maintain a zero-dependency strategic sandbox (Director) that operates out-of-tree.
  * We avoid premature architecture in the runtime project repositories.