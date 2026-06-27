# ADR/ARD Rules for Change Reviews

Use ADR (Architecture Decision Record). If the team uses ARD terminology, treat ARD as equivalent for this process.

## When ADR/ARD is Mandatory

Create ADR/ARD section when change includes:

- Architecture or design decision
- Contract/API evolution
- Data model/schema modification
- Security/performance trade-off
- Non-trivial refactor with impact risk

## Required ADR/ARD Structure

- Decision: what was chosen
- Context: constraints and forces
- Alternatives: options considered
- Trade-offs: pros/cons and risks
- Consequences: immediate and future impacts
- Status: proposed/accepted/superseded

## Decision Quality Rules

- Do not record only final choice; include alternatives.
- Include technical and operational consequences.
- Include rollback strategy when decision is reversible.
- Link decision to affected artifacts and tests.
