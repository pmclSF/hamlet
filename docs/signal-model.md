# Signal Model

Signals are the core abstraction in Hamlet.

All user-visible insights must be reducible to structured signals.

## Root model

The root artifact is `TestSuiteSnapshot`.

A snapshot represents Hamlet's current understanding of a repository at a point in time.

## TestSuiteSnapshot

A snapshot contains:

- Repository
- Frameworks
- TestFiles
- CodeUnits
- Signals
- Risk
- Ownership
- Policies
- Metadata

## RepositoryMetadata

Represents repository-level context.

Suggested fields:
- name
- root path
- languages
- package managers
- CI systems
- snapshot timestamp
- git commit sha

## Framework

Represents a detected test framework.

Suggested fields:
- name
- version
- type
- file count
- test count

Framework types:
- unit
- integration
- e2e
- performance
- visual
- contract
- property-based

## TestFile

Represents a discovered test file.

Suggested fields:
- path
- framework
- owner
- test count
- assertion count
- mock count
- snapshot count
- runtime stats
- linked code units
- applied signals

## CodeUnit

Represents a code element under test.

Suggested fields:
- name
- path
- kind
- exported
- complexity
- coverage
- linked test files
- owner

## RuntimeStats

Represents runtime behavior if artifacts are available.

Suggested fields:
- average runtime ms
- p95 runtime ms
- pass rate
- retry rate
- variance

## Signal

Every signal must include:

- type
- category
- severity
- confidence
- location
- owner
- explanation
- suggestedAction
- metadata

## Signal categories

- structure
- health
- quality
- migration
- governance

## Design constraints

### Signals must be:
- actionable
- explainable
- composable
- serializable
- stable enough for multiple surfaces

### Signals should not be:
- vague
- moralizing
- untraceable to evidence
- dependent on hidden logic

## Why signals matter

Signals allow Hamlet to unify:
- static analysis
- runtime evidence
- migration analysis
- policy checks
- risk scoring
