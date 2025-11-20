# GEMINI.md

## Project Overview

`agent-runner` is a meta-agent and orchestration layer designed to autonomously execute tasks by managing AI-based worker agents (e.g., coding assistants). It aims to minimize human intervention by orchestrating worker agents in a safe and reproducible manner within Docker sandboxes. After a task is completed, it generates a "memory" of the task as a Markdown file, which can be used by other agents or humans for context.

The architecture consists of three main components:
-   **Meta-agent (LLM):** The "brain" that plans tasks, decides on the next actions, and evaluates the results.
-   **AgentRunner Core:** The "hands" that execute the Meta-agent's decisions. It manages Docker sandboxes, executes worker CLIs, manages the task state, and generates Markdown notes.
-   **Worker Agents (CLI):** The "doers" that perform the actual work (e.g., coding, running tests, using git) inside a sandbox.

The system uses YAML for structured protocols and instructions between components and generates Markdown files as a persistent record of a task's lifecycle and outcomes.

## Building and Running

The project is a standard Go application.

### Building the application

To build the `agent-runner` binary, run:
```bash
go build ./cmd/agent-runner
```

### Running the application

The `agent-runner` reads a task configuration from a YAML file provided via standard input.

```bash
./agent-runner < task.yaml
```
A sample task configuration can be found in `sample_task_go.yaml`.

### Running tests

To run the project's tests, use the standard Go test command:
```bash
go test ./...
```

## Development Conventions

-   **Configuration:** Task configurations are defined in YAML files.
-   **Communication Protocol:** Communication between the Core and the Meta-agent is handled via YAML-based protocols.
-   **Task Output:** The results and "memory" of each task are stored in Markdown files, typically in the `.agent-runner/` directory.
-   **Sandboxing:** All worker actions are executed within a Docker container for isolation and reproducibility. The `sandbox/Dockerfile` defines the base environment for these workers.
-   **Modularity:** The project is organized into several internal packages (`core`, `meta`, `worker`, `note`) and a `pkg` directory for shared configuration, following common Go practices.

## Interaction Policy

-   **Language:** Please always communicate in Japanese.
