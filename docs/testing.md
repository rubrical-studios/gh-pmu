# Testing Configuration

## Test Projects

For integration testing, use these dedicated test projects instead of the main project (#11):

| Project | Number | Purpose |
|---------|--------|---------|
| Test Project 1 | **17** | Integration testing |
| Test Project 2 | **18** | Integration testing |

**IMPORTANT:** Do NOT create test issues or run integration tests against Project #11 (production).

## Running Tests

### Unit Tests
```bash
go test ./...
```

### Integration Tests
When testing commands that create/modify issues, use one of the test projects:
- Temporarily modify `.gh-pmu.yml` to point to project 17 or 18
- Or use a separate test config file

## Test Issues Created

Track test issues here for cleanup:
- (none currently)
