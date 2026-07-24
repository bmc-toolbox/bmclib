# Using AI Agents in `bmclib`

This document outlines how AI agents can be used to assist in the development, review, and maintenance of the `bmclib` repository. The goal is to leverage agents to improve code quality, ensure consistency, and provide clear guidance to contributors.

### Understanding the Codebase

Agents can quickly build an understanding of the repository's structure and control flow. Instead of manual searching, use an agent to:

*   **Discover Implementations:** Find which providers implement a specific `bmclib` interface.
    *   *Example Prompt: "Show me all providers that implement the `FirmwareInstaller` interface and trace how the `Install()` method is called."*
*   **Analyze Component Interaction:** Understand the relationship between internal packages, like how providers use the `redfishwrapper`.
    *   *Example Prompt: "Trace the call graph for `redfishwrapper.Client.Open()` to see how Redfish sessions are established and used by providers."*

### Reviewing Pull Requests

Agents can provide deep and insightful code reviews that go beyond surface-level checks.

*   **Pattern and Convention Analysis:** An agent can verify that a PR adheres to established architectural patterns. It can check, for example, that new features are exposed via the generic `bmc` interfaces and that provider-specific logic remains properly encapsulated.

*   **Dependency-Aware Review:** This is a key strength. An agent can analyze a PR's changes against the project's existing dependencies to prevent re-implementing functionality. This was evident in a recent review where an agent identified that new, manual Redfish calls could be replaced by using the existing `gofish` library.
    *   *Example Prompt: "This PR adds support for Redfish Certificate Management. Can you check if our `gofish` dependency already provides abstractions for this?"*

*   **Consistency Checks:** An agent can ensure that contributions use shared libraries correctly, such as verifying that vendor-specific strings are replaced with constants from `bmc-toolbox/common`.

### Guiding Contributions with Actionable Feedback

An agent's most valuable role is turning analysis into clear, actionable guidance for contributors.

*   **Generating Concrete Refactoring Plans:** Instead of a vague comment like "Please use gofish," an agent can produce a detailed, step-by-step plan with code examples.

*   **Proposing Architectural Improvements:** An agent can suggest improvements that make the codebase more robust and maintainable. For example, rather than just suggesting a provider use a dependency directly, it can recommend a more encapsulated approach:

    > "To improve encapsulation, the `redfishwrapper` should not expose the `gofish` client directly. Instead, it should expose methods that return the necessary `gofish` objects (e.g., `Systems()`, `Chassis()`, `Managers()`). This decouples the providers from the underlying client, making future changes easier."

By using agents in these capacities, we can ensure that `bmclib` remains a high-quality, maintainable, and extensible library.

### Contributing with Git

When contributing to `bmclib`, it's important to follow best practices for Git commits. This ensures a clean, readable, and maintainable commit history.

*   **Commit Hygiene:** All changes should be grouped into logical, atomic commits. Each commit should represent a single, self-contained change. This makes it easier to review, understand, and, if necessary, revert changes. Avoid large, monolithic commits that bundle unrelated changes together.

*   **Commit Messages:** Commit messages should be clear, concise, and written in the imperative mood (e.g., "Add support for X" instead of "Added support for X"). The message should explain the "why" of the change, not just the "what." A well-written commit message provides context that is invaluable for future maintenance and debugging.

