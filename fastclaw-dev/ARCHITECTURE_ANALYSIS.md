# FastClaw Architecture Analysis

This document provides a comprehensive architectural analysis of the FastClaw codebase, serving as the foundation for understanding its structure and guiding the development of the VEDA Agent Runtime within VEDA Core.

## Overall Architecture

FastClaw follows a modular, layered architecture designed for extensibility and multi-tenancy. The system is organized around several core principles:

1. **Separation of Concerns**: Distinct layers handle different responsibilities (gateway, agent, tools, storage, etc.)
2. **Multi-tenancy**: User and agent isolation through database-backed storage and workspace segregation
3. **Extensibility**: Plugin system, MCP support, and skill-based architecture for adding capabilities
4. **Sandboxing**: Secure execution environment for agent operations
5. **Event-driven Communication**: Internal messaging via message bus for loose coupling

The architecture consists of these primary layers:
- **Gateway Layer**: Runtime orchestrator handling connections, channels, and system services
- **Agent Layer**: ReAct (Reasoning and Acting) agent loop implementation
- **Tool Layer**: Executable capabilities available to agents
- **Storage Layer**: Persistent storage for configuration, state, and user data
- **Workspace Layer**: File system abstraction for agent-generated content
- **Skills Layer**: Modular capabilities that extend agent functionality

## Runtime Lifecycle

The runtime lifecycle is managed through the Gateway component (`internal/gateway/gateway.go`):

1. **Initialization**: 
   - Loads environment configuration
   - Initializes storage (SQLite/PostgreSQL)
   - Sets up workspace storage (local/S3/etc.)
   - Initializes message bus, channel manager, cron scheduler, webhook server, and plugin manager
   - Configures usage metering and quota enforcement

2. **Agent Lifecycle Management**:
   - Agents are loaded lazily on first authentication request via `UserSpaceFor()`
   - UserSpaces are cached and evicted based on idle time
   - Agents can be dynamically added/removed via manager APIs
   - Hot-reload capability through SIGHUP signal handling

3. **Request Processing Flow**:
   - Incoming requests (chat, API, webhook, etc.) are routed to appropriate handlers
   - For chat requests: UserSpace is retrieved/created, agent is obtained from the space
   - Agent processes the message through its ReAct loop (`HandleMessage`)
   - Response is sent back through the appropriate channel

4. **Shutdown**:
   - Graceful shutdown on SIGINT/SIGTERM
   - Stops all services in reverse order of initialization
   - Closes sandbox executors, task queue, plugin manager, etc.

## Execution Engine

The core execution engine is implemented in the Agent struct (`internal/agent/loop.go`), following the ReAct (Reasoning and Acting) pattern:

### Key Components:
1. **Context Builder**: Constructs the system prompt and context for the LLM
2. **Tool Registry**: Manages available tools and their execution
3. **Memory System**: Handles short-term (session) and long-term (MEMORY.md) memory
4. **Skills Loader**: Manages skill discovery and loading
5. **Hook System**: Allows interception of key points in the agent lifecycle
6. **Sandbox Integration**: Provides secure execution environment for tools
7. **MCP Integration**: Supports Model Context Protocol servers for extended capabilities

### Processing Flow:
1. Receive inbound message
2. Bind session context (sandbox, workspace, message routing)
3. Process through ReAct loop:
   - Generate LLM response with available tools
   - Parse and execute tool calls
   - Observe results and repeat until completion or max iterations
4. Apply post-processing (memory updates, goal continuation, etc.)
5. Return final response

### Key Loops:
- **Main Agent Loop**: `HandleMessage` → ReAct reasoning → Tool execution → Observation
- **Tool Execution Loop**: With safety checks, sandboxing, and result handling
- **Continuation Loop**: For goal-based autonomous behavior

## Agent Lifecycle

Agent lifecycle management is handled by the Manager (`internal/agent/manager.go`):

### Creation:
1. Resolved agent configuration is loaded from storage
2. Provider is selected (agent-specific or shared fallback)
3. Agent is constructed with:
   - Tool registry (built-in + skills + MCP + plugin tools)
   - Session manager
   - Memory system
   - Context builder
   - Hook registry
   - SDK engine (underlying ReAct implementation)
   - Cost tracker
   - Various stores (data, workspace, sandbox, etc.)

### Initialization:
- Skills are loaded from filesystem and object store
- MCP servers are connected and tools registered
- Hooks are initialized with logging
- Message tool is registered with split-reply capability
- Delegate task tool is registered for sub-agent spawning

### Runtime:
- Agents are retrieved from UserSpace on demand
- Each agent maintains its own state (memory, sessions, etc.)
- Agents can be dynamically added/removed without system restart
- Configuration changes trigger hot-reload via manager update methods

### Destruction:
- Agents are removed from manager map
- Resources are cleaned up through Close() methods on subcomponents
- References are cleared to allow garbage collection

## Scheduler

The scheduling system is implemented in the `cron` package:

### Components:
1. **Scheduler** (`internal/cron/scheduler.go`): Main scheduling interface
2. **Store Adapter** (`internal/cron/scheduler.go:cronStoreAdapter`): Database-backed job persistence
3. **Jobs**: Scheduled tasks stored in the database with metadata

### Features:
- Database-persisted jobs (no in-memory job list)
- Owner-based routing (jobs carry OwnerUserID for correct routing)
- Channel validation (scheduler checks if destination channel exists before firing)
- Automatic cleanup of missed jobs (configurable threshold)
- Integration with gateway's message bus for job execution

### Job Execution Flow:
1. Scheduler ticks at configured interval
2. Queries database for due jobs
3. For each job:
   - Validates channel availability via channel manager
   - Creates inbound message with job payload
   - Routes message through standard inbound processing pipeline
   - Updates job status (success/failure/reschedule)

## Event System

The event system combines multiple mechanisms for internal communication:

### Message Bus (`internal/bus`):
- Core publish/subscribe system for loose coupling
- Supports multiple implementations (in-memory, Redis-backed)
- Used for:
  - Agent-to-agent communication
  - System event broadcasting (heartbeat, goal completion, etc.)
  - Tool result propagation
  - Plugin communication

### Channels (`internal/channels`):
- Manages external communication channels (Telegram, Discord, Slack, Web, etc.)
- Handles inbound/outbound message routing
- Provides typing indicators and presence information
- Supports channel-specific features (markdown formatting, file sharing, etc.)

### WebSocket Events:
- Real-time updates for web clients via Server-Sent Events (SSE)
- Separate endpoints for chat streaming and event subscription
- Enables live updates for dashboards and monitoring tools

### Hook System (`internal/agent/hooks.go`):
- Synchronous interception points in agent lifecycle
- Points: Before/After Model Call, Before/After Tool Call, Post Turn
- Used for logging, metrics, policy enforcement, and custom behavior

## Plugin Architecture

The plugin system (`internal/plugin`) enables extension through external processes:

### Design:
- JSON-RPC based communication with plugin subprocesses
- Plugin discovery from configured directories
- Capability negotiation and versioning
- Sandboxed execution for security

### Integration Points:
1. **Tool Providers**: Plugins can register new tools available to agents
2. **Channel Adapters**: Plugins can add new communication channels
3. **System Hooks**: Plugins can participate in agent lifecycle events
4. **Resource Providers**: Plugins can provide access to external services

### Lifecycle:
1. Discovery: Plugin manager scans configured directories at startup
2. Validation: Plugins must implement required JSON-RPC methods
3. Activation: Plugins are started and registered with appropriate subsystems
4. Communication: JSON-RPC over stdin/stdout with the plugin process
5. Termination: Plugins are stopped during shutdown

## Skills/Capabilities System

Skills represent modular units of capability that can be added to agents:

### Structure:
- Each skill resides in its own directory under `skills/` or `agents/<id>/agent/skills/`
- Contains a `SKILL.md` file with metadata and instructions
- May include supporting files (scripts, templates, etc.)
- Can declare environment variables for API keys/configuration

### Loading Process:
1. SkillsLoader scans configured skill directories
2. For each skill directory:
   - Reads SKILL.md for metadata (name, description, etc.)
   - Validates required fields
   - Loads any associated scripts or templates
   - Prepares environment variable injection
3. Skills are made available through the tool registry via the `load_skill` tool
4. Skills can be updated at runtime without agent restart

### Types:
- **Built-in Skills**: Distributed with the core application (code-runner, image-gen, etc.)
- **Installed Skills**: Added via CLI or dashboard from skill repositories
- **Agent-private Skills**: Installed specifically for a single agent
- **Global Skills**: Available to all agents in the system

### Execution:
- Skills are invoked through the agent's tool system
- Typically provide one or more tools that perform specific functions
- Can access agent context, memory, and workspace
- May require specific environment variables for external service access

## Session Management

Session management handles conversation state and history:

### Components:
1. **Session Manager** (`internal/session/manager.go`): Tracks active sessions
2. **Session Store** (`internal/session/store_adapter.go`): Persistence layer for sessions
3. **Session Data**: Chat history, context, and temporary state

### Features:
- Per-user, per-agent session isolation
- Automatic cleanup of idle sessions
- History preservation for context in subsequent turns
- Support for session-specific variables and state
- Integration with memory system for long-term retention
- Project-bound sessions for collaborative workspaces

### Flow:
1. Session created on first message in a conversation
2. Each turn updates session with new messages and state
3. Session data persists across turns for context
4. Sessions expire after configurable idle timeout
5. Expired sessions are cleaned up by background evictor

## Memory Integration Points

Memory system provides both short-term and long-term storage:

### Short-term Memory (Session-based):
- Conversation history for current context
- Temporary variables and state during a session
- Automatically managed by session system
- Limited to current conversation window

### Long-term Memory (`internal/agent/memory.go`):
- **MEMORY.md**: Persistent facts and knowledge extracted from conversations
- **USER.md**: User profile and preferences
- **Automatic Extraction**: Information extracted from conversations via heartbeat mechanism
- **Manual Updates**: Users can directly edit memory files
- **Context Injection**: Memory content included in system prompt for LLM

### Integration Points:
1. **Heartbeat System**: Periodically processes conversation to extract memories
2. **Context Builder**: Includes relevant memory in LLM prompts
3. **Memory Tools**: `memory_search` tool allows agents to query their memory
4. **Persistence Layer**: Backed by storage system (SQL/PostgreSQL or filesystem)
5. **Privacy Controls**: PII scrubbing capabilities for sensitive data

## LLM/Model Abstraction

The LLM abstraction layer provides provider-agnostic access to language models:

### Provider Interface (`internal/provider/provider.go`):
- Standardized methods for chat completion and streaming
- Support for various parameter configurations (temperature, max tokens, etc.)
- Usage tracking and metrics collection
- Error handling and retry logic

### Supported Providers:
- OpenAI (GPT series)
- Anthropic (Claude series)
- Ollama (local models)
- OpenRouter (aggregator service)
- Groq, DeepSeek, Mistral, and any OpenAI-compatible API

### Abstraction Features:
1. **Provider Selection**: Per-agent or system-wide provider configuration
2. **Model Mapping**: Agent-specific model overrides
3. **Parameter Handling**: Consistent parameter passing across providers
4. **Response Normalization**: Standardized response format regardless of provider
5. **Streaming Support**: Real-time token delivery for interactive experiences
6. **Tool Calling**: Unified interface for function/tool calling capabilities
7. **Vision Support**: Image understanding capabilities where available
8. **Cost Tracking**: Token usage monitoring and reporting

### Configuration:
- Provider credentials stored securely in environment/vault
- Model selection per-agent via configuration
- Fallback mechanisms for provider failures
- Rate limiting and quota enforcement per provider

## Tool Execution System

The tool system provides secure, extensible capabilities for agents:

### Tool Registry (`internal/tools/registry.go`):
- Central repository for all available tools
- Supports built-in, MCP, and plugin-sourced tools
- Manages tool execution context (sandbox, workspace, permissions)
- Provides metadata for LLM tool calling

### Tool Categories:
1. **Built-in Tools**: Core filesystem and utility operations
   - `read_file`, `write_file`, `list_dir`: File system operations
   - `exec`: Command execution with sandboxing options
   - `web_search`, `web_fetch`: Internet access capabilities
   - `memory_search`: Query agent's long-term memory
   - `message`: Send messages to users
   - `delegate_task`: Create sub-agents for parallel work
   - `skill_manifest_gate`: Controlled access to skill metadata
   - `cron`: Schedule future tasks
   - `timezone`: Manage timezone settings
   - `tts`: Text-to-speech conversion

2. **MCP Tools**: From connected Model Context Protocol servers
3. **Plugin Tools**: From registered plugin subprocesses
4. **Skill Tools**: Dynamically loaded from skill directories

### Security Features:
- **Sandboxing**: Optional Docker or E2B execution for untrusted code
- **Path Restrictions**: File operations limited to designated directories
- **Identity File Protection**: SOUL.md, IDENTITY.md, etc. protected from unauthorized access
- **Skill Manifest Protection**: Prevents IP theft of skill definitions
- **Admin Gating**: Certain operations restricted to agent owners/admins
- **Environment Isolation**: Skills receive only configured environment variables

### Execution Flow:
1. LLM requests tool execution via standard tool call format
2. Tool registry validates request and prepares execution context
3. Tool function executed with appropriate sandboxing/workspace settings
4. Results returned to LLM for further processing
5. Side effects (file changes, messages sent, etc.) persisted as appropriate

## Sandbox System

The sandbox provides secure execution isolation for potentially unsafe operations:

### Modes:
1. **Disabled**: All execution occurs on host system (development only)
2. **Optional** (Self-hosted default): Host shell is default, sandbox available via `exec(sandbox:true)`
3. **Enforced** (Hosted default): All `exec` and file operations require sandbox

### Backends:
1. **Docker**: Local container-based isolation
2. **E2B**: Cloud-based sandbox environments
3. **Boxlite**: Alternative cloud sandbox provider

### Integration Points:
1. **Executor Pool**: Pre-warmed pool of sandbox executors for efficiency
2. **Bind Session**: Per-session sandbox assignment at start of each turn
3. **Tool Wrapping**: Executable tools automatically routed through sandbox when required
4. **Fallback Handling**: Clear error messages when sandbox unavailable in enforced mode
5. **File Synchronization**: Workspace changes mirrored between host and sandbox

### Security Guarantees:
- Process isolation prevents host system compromise
- Network restrictions limit external communication
- Filesystem boundaries prevent unauthorized access
- Resource limits prevent denial-of-service
- Clean state between executions prevents cross-contamination

## Configuration System

Configuration follows a layered approach:

### Bootstrap Configuration (Environment Variables):
- Set at process startup, cannot be changed at runtime
- Examples: `FASTCLAW_PORT`, `FASTCLAW_STORAGE_TYPE`, `FASTCLAW_SANDBOX_BACKEND`
- Loaded via `internal/config/env.go`

### Runtime Configuration (Database-backed):
- Modifiable via dashboard or CLI without restart
- Stored in `configs` table with scoping (system, agent, user)
- Includes: provider settings, model parameters, tool configurations, feature flags
- Loaded via `internal/config/config.go`

### Configuration Hierarchy:
1. Environment variables (bootstrap)
2. System settings (database, scope=system)
3. Agent settings (database, scope=agent)
4. User settings (database, scope=user)
5. Runtime overrides (per-request context)

### Key Configuration Areas:
- **Providers**: LLM API keys, endpoints, model mappings
- **Tools**: Enabled/disabled status, configuration parameters
- **Sandbox**: Backend selection, enforcement mode, image selection
- **Channels**: External service credentials and configuration
- **Skills**: Installation sources, auto-update settings
- **Limits**: Rate quotas, token limits, request limits
- **Features**: Enable/disable optional functionality

## Dependency Graph

Based on import analysis, the core dependencies are:

### Core Dependencies (no circular dependencies):
```
github.com/fastclaw-ai/fastclaw/internal/config
github.com/fastclaw-ai/fastclaw/internal/store
github.com/fastclaw-ai/fastclaw/internal/workspace
github.com/fastclaw-ai/fastclaw/internal/agent
github.com/fastclaw-ai/fastclaw/internal/gateway
github.com/fastclaw-ai/fastclaw/internal/bus
github.com/fastclaw-ai/fastclaw/internal/channels
github.com/fastclaw-ai/fastclaw/internal/cron
github.com/fastclaw-ai/fastclaw/internal/plugin
github.com/fastclaw-ai/fastclaw/internal/provider
github.com/fastclaw-ai/fastclaw/internal/sandbox
github.com/fastclaw-ai/fastclaw/internal/scope
github.com/fastclaw-ai/fastclaw/internal/session
github.com/fastclaw-ai/fastclaw/internal/taskqueue
github.com/fastclaw-ai/fastclaw/internal/toolproviders
github.com/fastclaw-ai/fastclaw/internal/usage
github.com/fastclaw-ai/fastclaw/internal/users
github.com/fastclaw-ai/fastclaw/internal/webhook
github.com/fastclaw-ai/fastclaw/internal/privacy
github.com/fastclaw-ai/fastclaw/internal/mcp
```

### External Dependencies:
- **Database**: SQLite (default) or PostgreSQL
- **Redis**: Optional for channel leasing and message bus
- **Object Store**: Local filesystem or S3-compatible (AWS, MinIO, etc.)
- **SSH**: For certain tool operations (git, scp, etc.)
- **Docker/E2B/Boxlite**: For sandbox execution
- **LLM Providers**: Various APIs as configured

### Internal Layer Dependencies:
1. **Foundation**: config, store, workspace
2. **Services**: bus, channels, cron, webhook, users, taskqueue
3. **Core Logic**: agent, gateway, provider, sandbox, scope, session
4. **Extensions**: toolproviders, plugin, mcp, privacy
5. **Interfaces**: API endpoints, CLI commands, web UI

## Package Responsibilities

### agent/
- **Core agent logic**: ReAct loop, memory, skills, tools, hooks
- **Lifecycle management**: Agent creation, configuration, execution
- **Context building**: System prompt construction and management
- **Integration points**: Memory persistence, goal management, MCP

### api/
- **HTTP API servers**: REST and WebSocket endpoints
- **Request/response handling**: Validation, authentication, routing
- **Versioning**: API version management

### auth/
- **Authentication systems**: API key validation, user resolution
- **Permission checking**: Role-based access control
- **Session authentication**: Web session management

### bus/
- **Message passing**: Publish/subscribe system for internal communication
- **Multiple implementations**: In-memory and Redis-backed

### channels/
- **External communication**: Telegram, Discord, Slack, Web, etc.
- **Inbound/outbound routing**: Message direction and formatting
- **Channel-specific features**: Typing, reactions, media handling

### config/
- **Configuration management**: Environment loading, runtime settings
- **Validation**: Configuration correctness checking
- **Migration**: Schema updates for configuration tables

### cron/
- **Scheduling system**: Timed job execution and management
- **Database persistence**: Job storage and recovery
- **Channel validation**: Ensuring targets exist before execution

### daemon/
- **Background service management**: Process supervision and control
- **Platform-specific implementations**: Unix and Windows services

### gateway/
- **Runtime orchestration**: Main system coordinator
- **User space management**: Per-user agent and session handling
- **Service initialization**: Starting all subsystem components
- **Signal handling**: Graceful startup and shutdown

### mcp/
- **Model Context Protocol**: Client for connecting to MCP servers
- **Tool exposure**: Making server capabilities available as agent tools
- **Resource handling**: Managing document and resource access

### plugin/
- **Plugin system**: Discovering, loading, and managing external plugins
- **JSON-RPC communication**: Standard interface for plugin interaction
- **Lifecycle management**: Start, stop, and monitoring of plugins

### provider/
- **LLM abstraction**: Unified interface for various language model providers
- **Provider implementations**: Specific adapters for each LLM service
- **Usage tracking**: Token counting and metrics collection
- **Error handling**: Normalization of provider-specific errors

### privacy/
- **PII detection and scrubbing**: Identifying and removing personal information
- **Configurable rules**: Customizable patterns for different data types
- **Integration points**: Applied to logs, memory, and user data

### sandbox/
- **Execution isolation**: Secure environments for running untrusted code
- **Multiple backends**: Docker, E2B, Boxlite implementations
- **Resource management**: Pooling and lifecycle of sandbox instances
- **Security policies**: Filesystem, network, and process restrictions

### scope/
- **Settings management**: Hierarchical configuration system
- **Namespace organization**: Different setting categories (sandbox, hooks, etc.)
- **Scope resolution**: Determining which settings apply to which contexts

### session/
- **Conversation state management**: Tracking active user interactions
- **History persistence**: Storing and retrieving chat history
- **Context maintenance**: Preserving relevant information across turns

### taskqueue/
- **Background job processing**: Asynchronous task execution
- **Worker pools**: Configurable concurrency levels
- **Task prioritization**: Ordering and scheduling of work items
- **Result handling**: Success/failure reporting and retry logic

### toolproviders/
- **Tool discovery**: Finding and loading available tools
- **Standard interfaces**: Common contracts for tool implementations
- **Registry management**: Tracking available tools and their sources
- **Security screening**: Validating tools before registration

### usage/
- **Metering systems**: Tracking resource consumption (tokens, requests)
- **Billing integration**: Usage data for invoicing and quotas
- **Analytics**: Collection of usage statistics and trends

### users/
- **Account management**: User creation, authentication, and permissions
- **API key handling**: Generation, validation, and revocation of keys
- **Session association**: Linking API keys to user identities

### webhook/
- **Incoming webhooks**: Receiving and processing external HTTP callbacks
- **Event routing**: Distributing webhook events to appropriate handlers
- **Security validation**: Verifying webhook authenticity and integrity

### workspace/
- **File system abstraction**: Unified interface for different storage backends
- **Cloud storage support**: S3, local filesystem, and other providers
- **Path normalization**: Consistent handling across different systems
- **Metadata tracking**: Timestamps, sizes, and other file attributes

## Folder-by-Folder Analysis

### Root Level:
- `README.md`: High-level overview, features, architecture, deployment
- `CHANGELOG.md`: Version history and notable changes
- `LICENSE`: Licensing terms (source-available with restrictions)
- `Makefile`: Build automation and common development tasks
- `go.mod/go.sum`: Go module dependencies and versions
- `Dockerfile`: Containerization instructions
- Various install scripts (`install.sh`, `install.ps1`)

### cmd/:
- Command-line interface implementations
- Subcommands for different administrative functions
- Main entry point in `main.go`

### internal/:
- Core application library code (not meant for external consumption)
- Organized by concern as detailed in Package Responsibilities above

### docs/:
- Integration contracts and API documentation
- `upstream-api.md`: External API for third-party integrations
- `coding-agent-runtime.md`: Project runtime and preview system

### deploy/:
- Deployment configurations for different environments
- Docker-compose, Kubernetes manifests, and helper scripts

### plugins/:
- Directory for plugin discovery and loading
- Empty by default, populated at runtime

### previews/:
- Static assets for demonstration purposes
- Screenshots and example outputs

### scripts/:
- Utility scripts for release processes, development helpers

### skills/:
- Built-in skills distributed with the core application
- Each skill in its own directory with SKILL.md and supporting files

### tools/:
- External tool integrations and bridges
- Examples: Plugin bridges for external systems

### web/:
- Web frontend implementation
- Next.js-based user interface for dashboard and controls

### workspace/:
- Default workspace directory for agent file operations
- Created at runtime, not checked into version control

## Internal Interfaces

Key interfaces that enable loose coupling and testability:

### Storage Layer:
- `store.Store`: Core database operations (CRUD, transactions, migrations)
- `workspace.Store`: Abstracted file system operations
- `session.SessionStore`: Persistence for chat sessions
- `usage.Meter`: Token and request counting
- `usage.QuotaStore`: Enforcement of usage limits
- `goal.Store`: Persistence for goal-related state

### Communication Layer:
- `bus.MessageBus`: Publish/subscribe messaging system
- `channels.Manager`: External communication channel management
- `plugin.Manager`: Plugin lifecycle and communication

### Abstraction Layers:
- `provider.Provider`: LLM service abstraction for language model interactions
- `sandbox.Executor`: Sandboxed execution environments
- `scope.Setting`: Hierarchical configuration access
- `toolproviders.Registry`: Tool discovery and registration

### Agent Components:
- `agent.ContextBuilder`: System prompt and context construction
- `agent.Memory`: Long-term memory storage and retrieval
- `agent.Sessions`: Conversation state management
- `agent.HookRegistry`: Lifecycle interception points
- `agent.SkillsLearner`: Automatic skill improvement system

### External Integrations:
- `webhook.Server`: Incoming HTTP webhook handling
- `mcp.Manager`: Model Context Protocol client connections

## Public Interfaces

### HTTP API (`/v1/*`):
Documented in `docs/upstream-api.md`:
- **Authentication**: Bearer token-based (API keys)
- **Endpoints**:
  - `POST /v1/chat/completions`: Main chat interface (OpenAI-compatible)
  - `GET /v1/agents`: List available agents
  - `POST /v1/users`: Provision end-users (optional)
  - `GET /v1/usage`: Token usage reporting
  - `PUT/GET/DELETE /v1/quota`: Usage quota management
- **WebSocket**: Real-time event streaming via SSE
- **Error Handling**: Standardized JSON error responses

### Coding-Agent Runtime API:
Documented in `docs/coding-agent-runtime.md`:
- **Project Management**: Creation and configuration of projects
- **Runtime Control**: Start, sleep, wake, and termination of development servers
- **Preview Generation**: URL generation for live previews
- **Log Access**: Retrieval of development server logs
- **Template Management**: Registration and selection of project templates

### CLI Interface:
- `fastclaw`: Main command with subcommands for:
  - Agent management (`agents`)
  - API key management (`apikey`)
  - Channel configuration (`channels`)
  - Plugin management (`plugin`)
  - Provider configuration (`provider`)
  - Sandbox configuration (`sandbox`)
  - Skill management (`skill`)
  - And many others

### Event System:
- Internal message bus for loose coupling between components
- WebSocket/SSE endpoints for real-time client updates
- Webhook system for outgoing notifications

## Extension Points

### 1. Plugin System:
- **Location**: `internal/plugin/`
- **Mechanism**: JSON-RPC subprocess communication
- **Capabilities**: Add new tools, channels, hooks, and system modifications
- **Registration**: Automatic discovery from configured directories
- **Isolation**: Runs in separate processes for security

### 2. Model Context Protocol (MCP):
- **Location**: `internal/mcp/`
- **Mechanism**: Standardized protocol for LLM-tool interaction
- **Capabilities**: Extend agent capabilities with external data sources and tools
- **Registration**: Configured via agent settings
- **Standardization**: Follows MCP specification for interoperability

### 3. Skills System:
- **Location**: `internal/agent/skills_learner.go` and skill directories
- **Mechanism**: Directory-based skill discovery with metadata files
- **Capabilities**: Add new agent capabilities through documented procedures
- **Distribution**: Shared via skill repositories (ClawHub, skills.sh, etc.)
- **Versioning**: Supports updates and dependency management

### 4. Tool System:
- **Location**: `internal/tools/`
- **Mechanism**: Registry-based tool discovery and execution
- **Capabilities**: Add new functions available to agents via LLM tool calling
- **Sources**: Built-in, MCP, plugin, or skill-provided
- **Security**: Sandboxing and permission controls

### 5. Channel System:
- **Location**: `internal/channels/`
- **Mechanism**: Adapter-based communication protocol implementation
- **Capabilities**: Add new communication platforms (Telegram, Discord, etc.)
- **Registration**: Database-backed with admin UI/CLI management
- **Bidirectional**: Support for both inbound and outbound communication

### 6. Provider System:
- **Location**: `internal/provider/`
- **Mechanism**: Interface-based LLM provider abstraction
- **Capabilities**: Add support for new language model services
- **Registration**: Configuration-based selection per agent or system
- **Standardization**: OpenAI-compatible interface with extensions

### 7. Hook System:
- **Location**: `internal/agent/hooks.go`
- **Mechanism**: Callback registration at key lifecycle points
- **Capabilities**: Insert custom logic at agent lifecycle events
- **Points**: Before/After model calls, before/after tool calls, post-turn
- **Ordering**: Defined execution sequence for multiple hooks

### 8. Storage System:
- **Location**: `internal/store/` and `internal/workspace/`
- **Mechanism**: Interface-based storage abstraction
- **Capabilities**: Add new persistence backends (databases, object stores)
- **Configuration**: Environment variable selection at startup
- **Consistency**: ACID transactions where applicable

## Architectural Strengths

### 1. Modularity and Separation of Concerns
- Clear boundaries between system components
- Well-defined interfaces reduce coupling
- Independent development and testing of subsystems
- Easy replacement of implementations (e.g., swapping storage backends)

### 2. Extensibility Framework
- Multiple, well-documented extension mechanisms
- Plugin system for arbitrary functionality
- MCP for standardized tool integration
- Skills for domain-specific capabilities
- Tool system for functional extensions

### 3. Security-Focused Design
- Defense-in-depth approach to security
- Sandboxing for untrusted code execution
- Principle of least privilege for file and system access
- Secure defaults with opt-in for less secure modes
- Identity and skill manifest protection against exfiltration

### 4. Multi-tenancy Foundation
- Robust user and agent isolation
- Per-user data segregation (USER.md, MEMORY.md)
- Resource metering and quota enforcement
- Support for both self-hosted and multi-tenant SaaS deployments

### 5. Observability and Debugging
- Comprehensive logging throughout the system
- Metrics collection for usage and performance
- Debugging tools and introspection capabilities
- Clear error messages and failure modes

### 6. Configuration Flexibility
- Environment-based bootstrapping for deployment flexibility
- Runtime configuration for operational adjustments
- Hierarchical scoping for inheritance and overrides
- Hot-reload capabilities for minimal disruption

### 7. Technology Choices
- Modern Go ecosystem with strong standard library
- Database abstraction for storage flexibility
- Modular design allowing technology swaps
- Active maintenance and community support

## Architectural Weaknesses

### 1. Complexity Surface Area
- Large number of moving parts increases cognitive load
- Multiple extension mechanisms can lead to integration complexity
- Distributed system characteristics introduce failure modes
- Steep learning curve for new contributors

### 2. Performance Considerations
- Message bus serialization can become bottleneck under high load
- Database round-trips for frequent operations may impact latency
- Serialization/deserialization overhead in plugin/MCP communication
- Memory usage growth over long-running agents

### 3. Consistency Challenges
- Eventual consistency in distributed components
- Potential race conditions in concurrent access patterns
- Cache invalidation complexity in multi-instance deployments
- State synchronization challenges between components

### 4. Operational Overhead
- Multiple external dependencies (database, redis, object store, sandbox backends)
- Monitoring and alerting complexity for distributed system
- Backup and disaster recovery procedures for multiple data stores
- Version compatibility management across components

### 5. Scaling Limitations
- Vertical scaling limits of single gateway instance
- Database connection limits under high concurrency
- Message bus throughput constraints
- Session storage memory pressure with many active users

### 6. Developer Experience
- Extensive codebase requires significant onboarding time
- Debugging distributed issues can be challenging
- Testing complexity due to integration points
- Documentation gaps for internal APIs and extension points

## Scalability Considerations

### Horizontal Scaling:
1. **Stateless Components**: Gateway instances can be scaled behind load balancer
2. **Shared State**: Database, Redis, and object store provide shared persistence
3. **Session Affinity**: Not required due to externalized session storage
4. **Worker Distribution**: Task queue workers can be scaled independently
5. **Read Replicas**: Database read operations can be scaled with replicas

### Vertical Scaling:
1. **Resource Allocation**: CPU and memory allocation per instance based on load
2. **Connection Pooling**: Database and Redis connection pooling
3. **Batch Operations**: Efficient bulk operations where applicable
4. **Caching Strategies**: Appropriate caching for frequently accessed data

### Bottleneck Mitigation:
1. **Database Optimization**:
   - Proper indexing on frequently queried columns
   - Connection pooling to manage database connections
   - Read replicas for read-heavy workloads
   - Query optimization for common access patterns

2. **Message Bus Throughput**:
   - Partitioning strategies for high-volume topics
   - Consumer group scaling for parallel processing
   - Message batching where applicable
   - Dead letter queues for error handling

3. **Memory Management**:
   - Efficient data structures for in-memory state
   - Garbage collection tuning for workload characteristics
   - Object pooling for frequently allocated objects
   - Memory pressure monitoring and alerts

4. **I/O Operations**:
   - Asynchronous operations where possible
   - Buffered I/O for file and network operations
   - Connection reuse and pooling
   - Asynchronous processing pipelines

### Deployment Patterns:
1. **Microservices**: Decompose services for independent scaling
2. **Service Mesh**: Traffic management and observability (Istio/Linkerd)
3. **Caching Layers**: Redis for frequently accessed computed values
4. **CDN Integration**: Static asset delivery for web interface
5. **Database Sharding**: Horizontal partitioning for massive scale

### Cloud-Native Features:
1. **Health Checks**: Liveness and readiness probes for orchestration
2. **Metrics Export**: Prometheus-compatible endpoints for monitoring
3. **Logging Structured Output**: JSON logging for log aggregation systems
4. **Distributed Tracing**: OpenTelemetry integration for request tracing
5. **Configuration Management**: Externalized configuration (Consul, Etcd, Vault)

## Data Flow Analysis

### 1. Chat Message Flow:
```
User Message → Gateway Handler → UserSpace Lookup → Agent Retrieval
     ↓
Agent HandleMessage → Bind Session (sandbox, workspace, context)
     ↓
ReAct Loop: LLM → Tool Calls → Observation → (Repeat)
     ↓
Response Processing → Memory Updates → Goal Continuation → Response Delivery
     ↓
Channel-Specific Formatting → External API Call → User Reception
```

### 2. File Operation Flow:
```
Tool Call (read_file/write_file) → Tool Registry Validation
     ↓
Sandbox Check (if required) → Executor Acquisition
     ↓
Workspace Path Resolution (session/project scoping)
     ↓
Storage Layer Dispatch (local/S3/etc.) → Actual File Operation
     ↓
Result Packaging → Return to LLM for Further Processing
```

### 3. Skill Loading Flow:
```
Agent Initialization → SkillsLoader Directory Scan
     ↓
SKIRT.md Parsing → Metadata Extraction and Validation
     ↓
Tool Registration (via load_skill tool mechanism)
     ↓
Runtime Availability → LLM Can Invoke Skill Tools
     ↓
Skill Execution → Access to Agent Context and Resources
```

### 4. Plugin Communication Flow:
```
Plugin Discovery → JSON-RPC Connection Establishment
     ↓
Capability Negotiation → Method Registration
     ↓
Remote Procedure Calls → Result Serialization/Deserialization
     ↓
Error Handling → Connection Lifecycle Management
```

### 5. MCP Integration Flow:
```
MCP Server Connection → Tool/Resource Enumeration
     ↓
Capability Registration → Standard Tool Interface Exposure
     ↓
Standard Tool Invocation → MCP Protocol Translation
     ↓
Server Execution → Result Return Through Standard Interface
```

## VEDA Core Architecture Mapping

Based on this analysis, the following mappings to VEDA Core architecture are recommended:

### Direct Adoptions (Keep as-is):
1. **Plugin System**: Well-designed, secure, and extensible
2. **MCP Integration**: Standardized approach for external tool integration
3. **Tool Registry**: Centralized, secure tool management with clear interfaces
4. **Hook System**: Clean extension points for lifecycle interception
5. **Configuration Hierarchy**: Flexible, layered approach suitable for enterprise
6. **Storage Abstraction**: Clean separation enabling multiple backend options
7. **Workspace Isolation**: Secure file system boundaries with project scoping

### Recommended Adaptations:
1. **Agent Lifecycle**: 
   - Adapt for VEDA's specific agent creation and management patterns
   - Consider integrating with VEDA's agent registry and lifecycle hooks
   - Maintain core ReAct loop but adapt context building for VEDA's needs

2. **Memory System**:
   - Align with VEDA's memory architecture and knowledge graph approach
   - Potentially replace MEMORY.md/USER.md with VEDA's storage primitives
   - Preserve the concept of short-term (session) and long-term memory

3. **Scheduler**:
   - Evaluate against VEDA's job scheduling and cron requirements
   - Consider direct integration or adaptation based on feature overlap
   - Maintain database persistence and owner-based routing concepts

4. **Event System**:
   - Map to VEDA's event-driven architecture patterns
   - Consider using VEDA's message bus if available, otherwise adapt patterns
   - Preserve loose coupling and extensibility principles

5. **Security Model**:
   - Adapt sandboxing approach to VEDA's security requirements
   - Maintain principle of least privilege and defense-in-depth
   - Adjust identity and credential management to VEDA's IAM system

### Components to Re-evaluate:
1. **LLM Provider Abstraction**:
   - Determine if VEDA has its own preferred LLM abstraction layer
   - May need to adapt or replace with VEDA's standard approach
   - Preserve the concept of pluggable providers if aligned with VEDA goals

2. **Channel System**:
   - Assess against VEDA's communication and integration requirements
   - May need to replace with VEDA's standard messaging mechanisms
   - Keep the concept of extensible communication channels if relevant

3. **Web Interface**:
   - Likely replace with VEDA's standard UI framework
   - Preserve UI/UX concepts but implement with VEDA's technology stack
   - Maintain the separation between API and presentation layers

### Implementation Approach:
1. **Core Engine**: Adapt the ReAct agent loop as the foundation
2. **Integration Layer**: Build adapters to VEDA's services and APIs
3. **Extension Points**: Map VEDA's extension mechanisms to these patterns
4. **Security Layer**: Align with VEDA's security model and requirements
5. **Observability**: Integrate with VEDA's monitoring and logging systems
6. **Configuration**: Map to VEDA's configuration management approach

This architecture provides a solid foundation that aligns well with modern AI agent system principles while offering clear extension points for VEDA-specific requirements and conventions.