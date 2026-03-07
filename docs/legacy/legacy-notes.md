# Legacy Notes

This file describes the relationship between the pre-V3 Hamlet codebase and the V3 architecture.

## Intent

V3 is a refactor, not a discard-and-rebuild effort.

Existing migration functionality should be preserved where possible and gradually mapped into:

- internal/migration
- internal/signals
- internal/reporting
- cmd/hamlet

## Migration strategy

1. Freeze docs and architecture
2. Introduce canonical models and signals
3. Build `hamlet analyze`
4. Preserve old behavior via adapters where necessary
5. Move legacy migration logic behind the new architecture gradually

## Rule

Do not delete working code until:
- a replacement exists
- tests cover the replacement
- architectural ownership is clear
