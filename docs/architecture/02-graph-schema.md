# Graph Schema

> **Status:** Superseded by [16-unified-graph-schema.md](16-unified-graph-schema.md)
>
> This document originally described the graph data model. It has been replaced by the unified graph schema document, which was hardened in the schema hardening pass to reflect the actual implementation.

**Canonical reference:** [16-unified-graph-schema.md](16-unified-graph-schema.md)

The unified graph schema defines:
- **20 node types** across 6 families (system, validation, behavior, environment, execution, governance)
- **15 edge types** with confidence scoring and evidence types
- **5 evidence types** (static_analysis, convention, inferred, manual, execution)
- **Deterministic serialization** via `MarshalJSON`/`UnmarshalJSON`
- **10-stage build pipeline** constructing the graph from snapshot data

Types removed during schema hardening (no longer in implementation):
- Node types: package, service, generated_artifact, external_service, fixture, helper
- Edge types: test_uses_fixture, test_uses_helper, fixture_imports_source, helper_imports_source, validates, test_exercises, depends_on_service
