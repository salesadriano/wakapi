---
name: documentation-sync-governance
description: Ensure every implemented, corrected, described, or modeled change generates complete review documentation with traceability, PRD alignment, ADR/ARD decision records, Mermaid diagrams, validation evidence, and handover clarity for third parties. Reusable across projects.
---

# Documentation Sync Governance

## Overview

This skill enforces a documentation-first and review-first quality gate: no change is complete until the related review document is complete, detailed, traceable, and understandable by a third party with no prior context.

It is project-agnostic and can be applied to any codebase by mapping local conventions.

## Goals

- Keep implementation and documentation consistent.
- Prevent knowledge drift after fixes or quick patches.
- Standardize traceability between requirement, code, tests, and decisions.
- Make onboarding and audits easier.
- Force complete review artifacts with mandatory architecture and process visualization.
- Ensure business intent (PRD) and technical decisions (ADR/ARD) are explicitly linked.

## Universal Rule

Any change in behavior, contract, structure, operation, requirement, or business rule MUST produce complete review documentation in the same delivery cycle.

If documentation cannot be updated in the same cycle, the delivery must be marked as blocked with explicit risk and a recovery plan.

## Mandatory Deliverable

Every change must generate or update one review document in Markdown with, at minimum:

- Context and objective
- Business alignment and PRD impact
- Technical decision log (ADR/ARD)
- Scope and affected files/modules
- Validation evidence and quality gates
- Risks, impact, rollback, and mitigation
- At least one Mermaid diagram
- Third-party handover notes (what changed, why, how to validate)

## Supported Change Types

Apply this skill to:

- New features
- Bug fixes
- Refactors with behavioral impact
- API contract changes (OpenAPI, TypeSpec, GraphQL, gRPC, etc.)
- Data model/schema changes
- Security changes
- Performance changes
- CI/CD and operational changes
- Test strategy and quality gate changes

## Inputs Required

Before applying the workflow, collect:

- Change summary
- Impacted modules/components
- Impacted interfaces/contracts
- User/business impact
- Validation evidence available
- Documentation locations used by the project

## Project Mapping (Portability Layer)

Map the current repository to these logical locations:

- Product/Business docs: for scope, requirements, user stories
- Technical docs: architecture, APIs, runbooks, ADRs
- Change log/review docs: release notes, review records, migration notes
- Test evidence docs: quality reports, test execution summaries

Use default paths only if the project has no explicit convention.

Suggested defaults:

- docs/
- docs/architecture/
- docs/api/
- docs/runbooks/
- docs/changelog/
- review/

Review document naming recommendation:

- YYYY-MM-DD-HHMM-short-slug.md

## Mandatory Workflow

### Step 1 - Classify the change

Tag the change with one or more labels:

- feat
- fix
- refactor
- docs-impact
- api-impact
- data-impact
- security-impact
- ops-impact
- prd-impact
- adr-impact

### Step 2 - Build the doc impact matrix

Create/update a matrix with at least:

- What changed
- Why it changed
- Affected artifact(s)
- Required documentation update(s)
- Owner
- Status

Use the template in references/traceability-matrix-template.md.

### Step 3 - Capture business intent (PRD rules)

For each change, register PRD-oriented information:

- Problem being solved
- Stakeholders impacted
- User/business outcomes expected
- Scope in/out
- Acceptance criteria and measurable success

Apply rules from references/prd-rules.md.

### Step 4 - Capture technical decisions (ADR/ARD rules)

For each non-trivial change, include an ADR/ARD block:

- Decision statement
- Context and constraints
- Alternatives considered
- Trade-offs
- Consequences

Apply rules from references/adr-ard-rules.md.

### Step 5 - Update documents in the same cycle

At minimum, update all impacted items:

- Requirement/spec artifacts (if business behavior changed)
- Technical artifacts (if design/flow/contracts changed)
- Operational artifacts (if setup/run/deploy changed)
- Test artifacts (if tests/scenarios/coverage gates changed)

If the delivered change completes an activity already listed in an existing follow-up section such as Next steps, Próximos passos, Próximas etapas, Pending work, or equivalent, update that source document in the same cycle and mark the matching activity as completed using the repository's local convention.

### Step 6 - Add validation evidence

Every documented change must include:

- Validation command(s)
- Observed result
- Scope verified

Containerized quality evidence policy (recommended default):

- Prefer: docker compose run --rm <service> <test-command>
- Or: docker compose exec <service> <test-command>
- Do not accept host-only test evidence when the project is containerized

### Step 7 - Add mandatory Mermaid diagrams

Each review document must include at least one Mermaid diagram from one of these categories:

- Architecture or dependency flow
- Business/process flow
- Sequence of a critical interaction
- State transition when applicable

Apply syntax and selection rules from references/mermaid-review-patterns.md.

### Step 8 - Publish complete review record

Create one review/change note containing:

- Context and objective
- Files changed (implementation + docs)
- Decision summary and alternatives considered
- Validation evidence
- Risks and rollback plan
- Next steps

When one of those next steps was previously planned in another document, explicitly reference the source artifact and record that the corresponding item was closed.

Use references/review-complete-template.md.

### Step 9 - Third-party handover verification

Confirm that a new engineer can understand and validate the change without additional context:

- What changed
- Why it changed
- How to test and verify
- What can fail and how to rollback
- Which docs/contracts were updated

Use references/third-party-handover-checklist.md.

### Step 10 - Completion gate

Mark task complete only when all conditions are true:

- Implementation updated
- Documentation updated
- Traceability matrix updated
- Validation evidence attached
- Review record published
- PRD impact explicitly documented
- ADR/ARD decision documented (when applicable)
- Mermaid diagram included
- Third-party handover checklist passed
- Any matching item from a prior Next steps or similar section was marked as completed in its source document (when applicable)

## Definition of Done (Docs Sync)

A delivery is done only if:

- No behavior change is undocumented.
- No API/data change is undocumented.
- Test strategy and evidence are documented.
- Consumer-facing impact is explicit.
- Rollback and risk notes exist for non-trivial changes.
- PRD impact is explicit.
- ADR/ARD is explicit for non-trivial technical decisions.
- At least one Mermaid diagram exists in the review document.
- Third-party handover is possible without synchronous explanation.

## Audit Checklist

Use references/docs-sync-checklist.md before closing any task.

## Anti-Patterns (Do Not Allow)

- "Small fix" without docs update.
- API changes without contract/version note.
- Refactor changing behavior with no requirement update.
- Missing quality evidence.
- Review note without risk/rollback.
- Review note without diagram.
- Review note without PRD impact.
- Review note without ADR/ARD where applicable.
- Completing a previously planned activity without closing it in the originating Next steps or equivalent section.
- Review that depends on oral context to be understood.

## Output Format Guidance

When this skill is invoked, produce:

1. A documentation impact summary table.
2. The exact list of files to update.
3. PRD impact block.
4. ADR/ARD decision block.
5. Mermaid diagram block.
6. Validation evidence block.
7. Third-party handover checklist result.
8. A completion gate result (PASS/FAIL with reasons).
9. A note identifying any prior Next steps or similar item that was closed, including the source document updated.

## Required References

This skill must use these reference files:

- references/docs-sync-checklist.md
- references/traceability-matrix-template.md
- references/prd-rules.md
- references/adr-ard-rules.md
- references/mermaid-review-patterns.md
- references/review-complete-template.md
- references/third-party-handover-checklist.md

## Reuse in Other Projects

To reuse this skill in another repo:

1. Copy this folder to .github/agents/skills/documentation-sync-governance/
2. Adjust only the Project Mapping section for local paths.
3. Keep workflow, templates, and completion gate unchanged.
4. If needed, adapt only naming conventions and documentation paths.
