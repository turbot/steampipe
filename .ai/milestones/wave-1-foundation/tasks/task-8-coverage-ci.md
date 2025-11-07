# Task 8: Coverage & CI Integration

## Context
Currently there is NO code coverage tracking. We need to add coverage reporting to CI/CD and track progress.

## Goal
Integrate code coverage reporting into CI/CD pipeline and establish coverage tracking.

## Files to Modify
- `.github/workflows/11-test-acceptance.yaml` - Add coverage to existing test workflow
- `Makefile` - Add coverage targets

## Tasks

### 1. Add Coverage to Makefile
```makefile
.PHONY: test-coverage
test-coverage:
	go test -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

.PHONY: test-coverage-report
test-coverage-report:
	go test -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out

.PHONY: test-coverage-check
test-coverage-check:
	@go test -coverprofile=coverage.out -covermode=atomic ./... > /dev/null
	@coverage=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Total coverage: $$coverage%"; \
	if [ $$(echo "$$coverage < 15" | bc -l) -eq 1 ]; then \
		echo "Coverage $$coverage% is below 15% target"; \
		exit 1; \
	fi
```

### 2. Update GitHub Actions Workflow
Add to `.github/workflows/11-test-acceptance.yaml`:

```yaml
- name: Run unit tests with coverage
  run: |
    go test -coverprofile=coverage.out -covermode=atomic ./...
    go tool cover -func=coverage.out

- name: Upload coverage to Codecov
  uses: codecov/codecov-action@v3
  with:
    files: ./coverage.out
    flags: unittests
    fail_ci_if_error: true

- name: Check coverage threshold
  run: |
    coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    echo "Total coverage: $coverage%"
    if (( $(echo "$coverage < 15" | bc -l) )); then
      echo "Coverage $coverage% is below 15% minimum"
      exit 1
    fi
```

### 3. Add Coverage Badge to README
In README.md, add:
```markdown
[![Coverage](https://codecov.io/gh/turbot/steampipe/branch/develop/graph/badge.svg)](https://codecov.io/gh/turbot/steampipe)
```

### 4. Create codecov.yml Config
```yaml
# .codecov.yml
coverage:
  status:
    project:
      default:
        target: 15%           # Wave 1 target
        threshold: 1%         # Allow 1% drop
    patch:
      default:
        target: 40%           # New code should be well tested

comment:
  require_changes: true
```

## Success Criteria
- [ ] Makefile has coverage targets
- [ ] CI runs coverage on every PR
- [ ] Coverage reports upload to Codecov
- [ ] Coverage badge in README
- [ ] CI fails if coverage drops below threshold
- [ ] Coverage report accessible in PR

## Testing Your Work
```bash
# Test locally
make test-coverage
make test-coverage-report
make test-coverage-check

# View HTML report
open coverage.html

# Test threshold check
# (should pass with Wave 1 complete)
```

## Dependencies
Requires Tasks 2-7 complete (must have tests to measure!)

## Time Estimate
2 hours

## Notes
- Coverage threshold starts at 15% for Wave 1
- Will increase threshold for each wave
- Codecov is free for open source

## Command
```bash
# WAIT for Tasks 2-7 to complete first!
claude
# "Please complete task-8-coverage-ci.md from .ai/milestones/wave-1-foundation/tasks/"
```
