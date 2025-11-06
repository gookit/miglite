---

description: "Task list for implementing MigLite Tool"
---

# Tasks: Implement MigLite Tool

**Input**: Design documents from `/specs/001-impl-miglite-tool/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, quickstart.md

**Tests**: The feature specification requires test coverage. Test-First approach is non-negotiable per project constitution.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Single project**: `pkg/`, `cmd/`, `tests/` at repository root
- Paths shown below assume single project - adjust based on plan.md structure

<!-- 
  ============================================================================
  The tasks below are based on:
  - User stories from spec.md (with their priorities P1, P2, P3...)
  - Feature requirements from plan.md
  - Entities from data-model.md
  - Implementation details from research.md
  
  Tasks MUST be organized by user story so each story can be:
  - Implemented independently
  - Tested independently
  - Delivered as an MVP increment
  ============================================================================-->

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Create project structure per implementation plan in pkg/ directory
- [ ] T002 Initialize Go project dependencies (github.com/gookit/goutil/cflag, github.com/gookit/goutil/testutil/assert)
- [ ] T003 [P] Configure Go linting and formatting tools (gofmt, golint)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T004 [P] Setup database connection infrastructure in pkg/database/
- [ ] T005 [P] Implement configuration management (miglite.yaml, .env, DATABASE_URL) in pkg/config/
- [ ] T006 [P] Create migration file structure and parsing logic in pkg/migration/
- [ ] T007 Create base Migration File and Migration Record models in pkg/migration/models.go
- [ ] T008 Configure error handling and logging infrastructure in pkg/common/
- [ ] T009 Setup CLI framework using github.com/gookit/goutil/cflag in pkg/cli/

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Basic Migration Execution (Priority: P1) 🎯 MVP

**Goal**: User can run database migration scripts by preparing SQL files and using miglite up command to execute migrations

**Independent Test**: Can prepare a simple SQL migration file, connect to a database, run miglite up command and after successful execution, the database schema should reflect the migration changes.

### Tests for User Story 1 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T010 [P] [US1] Unit test for migration file parsing in tests/unit/migration/parsing_test.go
- [ ] T011 [P] [US1] Integration test for database migration execution in tests/integration/migration_execution_test.go
- [ ] T012 [P] [US1] Unit test for configuration loading in tests/unit/config/config_test.go

### Implementation for User Story 1

- [ ] T013 [P] [US1] Implement migration file discovery in pkg/migration/discovery.go
- [ ] T014 [US1] Implement migration execution logic for UP operations in pkg/migration/executor.go
- [ ] T015 [US1] Implement migration status check in pkg/migration/status.go
- [ ] T016 [US1] Add CLI command for 'up' in cmd/miglite/commands/up.go
- [ ] T017 [US1] Add CLI command for 'status' in cmd/miglite/commands/status.go
- [ ] T018 [US1] Implement database migration tracking in pkg/database/tracker.go
- [ ] T019 [US1] Add validation and error handling for migration execution
- [ ] T020 [US1] Add logging for user story 1 operations

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Migration Rollback (Priority: P2)

**Goal**: User can rollback executed database migrations when a migration causes issues or needs to be reversed

**Independent Test**: By running a migration, then using miglite down command, the database should rollback to the state before the migration.

### Tests for User Story 2 ⚠️

- [ ] T021 [P] [US2] Unit test for migration rollback execution in tests/unit/migration/rollback_test.go
- [ ] T022 [P] [US2] Integration test for migration rollback in tests/integration/migration_rollback_test.go

### Implementation for User Story 2

- [ ] T023 [US2] Implement migration execution logic for DOWN operations in pkg/migration/executor.go
- [ ] T024 [US2] Add CLI command for 'down' in cmd/miglite/commands/down.go
- [ ] T025 [US2] Integrate rollback functionality with existing tracking (US1 components)

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Migration Creation (Priority: P3)

**Goal**: User can create new migration files using the tool with appropriate timestamp and UP/DOWN template structure

**Independent Test**: User runs miglite create command and gets a properly formatted SQL file with correct timestamp and -- Migrate:UP -- and -- Migrate:DOWN -- sections.

### Tests for User Story 3 ⚠️

- [ ] T026 [P] [US3] Unit test for migration file creation in tests/unit/migration/creation_test.go

### Implementation for User Story 3

- [ ] T027 [US3] Implement migration file creation logic in pkg/migration/creator.go
- [ ] T028 [US3] Add CLI command for 'create' in cmd/miglite/commands/create.go

**Checkpoint**: All user stories should now be independently functional

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T029 [P] Documentation updates in README.md and docs/
- [ ] T030 Code cleanup and refactoring across all packages
- [ ] T031 Performance optimization for handling large migration files
- [ ] T032 [P] Additional unit tests for edge cases in tests/unit/
- [ ] T033 Error handling hardening across all components
- [ ] T034 Run quickstart.md validation to ensure documentation matches implementation

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - May integrate with US1 components
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - Independent of US1/US2 components

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Models before services
- Services before endpoints/commands
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- All tests for a user story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
T010 [P] [US1] Unit test for migration file parsing in tests/unit/migration/parsing_test.go
T011 [P] [US1] Integration test for database migration execution in tests/integration/migration_execution_test.go
T012 [P] [US1] Unit test for configuration loading in tests/unit/config/config_test.go

# Launch all implementation tasks for User Story 1 together:
T013 [P] [US1] Implement migration file discovery in pkg/migration/discovery.go
T014 [US1] Implement migration execution logic for UP operations in pkg/migration/executor.go
T015 [US1] Implement migration status check in pkg/migration/status.go
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo
4. Add User Story 3 → Test independently → Deploy/Demo
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1
   - Developer B: User Story 2
   - Developer C: User Story 3
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence