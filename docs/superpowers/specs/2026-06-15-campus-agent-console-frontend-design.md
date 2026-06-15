# Campus Agent Console Frontend Design

## Goal

Build a warm-toned single-page console for Campus Agent MVP and serve it directly from the Go server. The page should demonstrate the existing backend capabilities without introducing a separate frontend build system.

## Scope

This design covers one browser entrypoint exposed from the existing Gin server:

- root page `/`
- health check display from `GET /healthz`
- chat interaction through `POST /api/v1/chat`
- async task creation through `POST /api/v1/tasks`
- task list through `GET /api/v1/tasks?user_id=...`
- task detail refresh through `GET /api/v1/tasks/:id`

Out of scope for this phase:

- login and authentication
- multi-page navigation
- frontend framework setup
- websocket or server-sent event updates
- visual editor for prompts or knowledge docs

## Product Shape

The page is a single-screen console with three functional zones:

1. Top bar
   - Campus Agent brand block
   - short product subtitle
   - backend health badge
   - current user ID input used by all requests
2. Main conversation area
   - conversation timeline with user and agent messages
   - chat composer for free-form requests
   - structured agent response cards showing intent, planned tasks, knowledge answer, and tool execution results
3. Right-side operations panel
   - async task creation form
   - task list for the current user
   - selected task detail card with status and result

Desktop uses a two-column layout. Mobile collapses to a vertical flow with the conversation first and operations below it.

## Visual Direction

The interface should feel like a modern campus operations desk rather than a generic SaaS dashboard.

- primary palette: warm orange, terracotta, sand, cream
- contrast color: deep brown for text and borders
- accent use: small green status only for healthy/success states
- background: layered warm gradients with subtle paper-like panels
- typography: use a serif-forward heading stack with a clean sans fallback for body text
- motion: light fade-in on load, hover lift for cards, pulsing status dots for running tasks

The page should avoid purple, neon gradients, heavy glassmorphism, and dark-mode-first styling.

## Information Architecture

### Header

The header gives immediate context and system status:

- product name `Campus Agent Console`
- one-line description focused on campus task automation
- health badge showing `online` or `offline`
- user ID field with a sensible default for local demo use

### Conversation Panel

The conversation panel is the primary entrypoint because it best expresses the agent workflow.

Each user message is shown as a compact warm bubble. Each agent response is shown as a card that can include:

- detected intent
- planned task list
- knowledge answer text
- tool execution result items

If the backend returns an error, show an inline error card in the conversation stream instead of a browser alert.

### Operations Panel

The operations panel provides direct control over async tasks and exposes the worker flow clearly.

Sections:

- `Create Task`
  - task ID input
  - task name input
  - submit button
- `Recent Tasks`
  - compact list of recent tasks for the current user
  - each row shows task name, status, and update time if available
- `Task Detail`
  - currently selected task ID
  - status, result, and timestamps
  - manual refresh action

## Interaction Design

### Page Boot

On initial load:

1. render the static shell
2. request `/healthz`
3. load the current user's tasks
4. if tasks exist, select the newest task and load its detail

If health check fails, the page still renders but marks the backend as offline and disables submit buttons until a later successful check or manual retry.

### Chat Flow

When the user submits a message:

1. append the user message to the timeline immediately
2. show a pending agent card
3. call `POST /api/v1/chat`
4. replace the pending state with structured response content

If the response contains knowledge answer text, render it as the main body of the agent card. If it contains task results, render them as labeled rows.

### Task Creation Flow

When the user submits a task:

1. validate that user ID, task ID, and task name are present
2. call `POST /api/v1/tasks`
3. prepend the created task to the task list
4. select the new task
5. start short polling its detail until status becomes `success` or `failed`

### Task Detail Flow

When a task row is selected:

1. call `GET /api/v1/tasks/:id`
2. render the detail card
3. if status is `pending` or `running`, poll every few seconds
4. stop polling once status is terminal

The page should maintain only one active polling loop at a time so task switching does not create duplicate requests.

## Error Handling

- invalid form input: show inline field-level hint and keep the page state intact
- failed health check: show offline badge and disable actions
- chat request failure: show error card in conversation panel
- task list/detail failure: show error text inside the relevant card area
- polling failure: stop the loop after a small number of consecutive failures and expose a manual refresh control

Error messages should be short, concrete, and visible in context.

## Accessibility and Responsiveness

- use semantic form controls and button labels
- maintain readable contrast on warm backgrounds
- support keyboard submission for chat and task forms
- ensure mobile layout works from roughly 360px width upward
- avoid animation patterns that block reading or interaction

## Technical Design

### File Layout

New files:

- `web/static/index.html`
- `web/static/styles.css`
- `web/static/app.js`

Server updates:

- register a root route for the page
- serve static assets under a predictable path such as `/static`

## Frontend Runtime

Use plain browser APIs only:

- `fetch` for API requests
- DOM template helpers for rendering cards and lists
- a small in-memory state object for current user ID, selected task ID, task list, and polling timer

No bundler, framework, or package manager is introduced in this phase.

## Backend Integration Assumptions

The page uses the current API contracts as-is:

- chat request body: `user_id`, `message`
- create task body: `id`, `user_id`, `task_name`
- task list filtered by `user_id`
- standard success and error envelopes from `pkg/response`

This phase does not require backend schema changes.

## Testing Strategy

Testing for this phase is intentionally lightweight but explicit:

- unit test server route registration for `/` and static asset serving
- verify `go test ./...` still passes
- manually open the page and validate:
  - health badge loads
  - chat request renders response card
  - task creation updates the sidebar
  - task detail polling stops on terminal status
  - mobile layout remains readable

## Implementation Notes

- keep the JS file focused on state and rendering; avoid embedding large HTML strings everywhere
- keep CSS organized by layout, component, and responsive overrides
- prefer a minimal number of DOM IDs and use data attributes where they improve clarity
- preserve existing Gin wiring patterns instead of creating a new app bootstrap path

## Acceptance Criteria

The phase is complete when:

- visiting `/` loads a styled single-page console
- the page is served by the existing Go server
- the page can perform chat, task creation, task list loading, and task detail refresh against the current backend
- the visual design is clearly warm-toned and non-generic
- no separate frontend toolchain is required
