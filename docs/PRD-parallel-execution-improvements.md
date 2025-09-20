# Product Requirements Document (PRD)
## Parallel Execution Improvements for pre-push

**Document Version:** 1.0  
**Date:** 2025-01-27  
**Status:** Draft  

### 1. Executive Summary

This PRD defines improvements to the pre-push tool's parallel execution engine to address two critical user experience issues:

1. **Ordered Output**: Steps should be displayed in the order they are declared in `project.yml`, not in execution order
2. **Parallel Execution on Failure**: When a step fails, other independent steps should continue running and report their results

### 2. Problem Statement

#### Current Issues

**Issue 1: Inconsistent Output Order**
- Steps are currently displayed in execution order (DAG topological order)
- Users expect to see steps in the same order they declared them in `project.yml`
- This creates confusion when debugging configuration issues

**Issue 2: Premature Termination on Failure**
- When a step fails, the executor immediately stops running other steps
- This prevents users from seeing the results of other checks that could have run independently
- Users lose valuable information about the overall state of their project

#### Example Scenario
```yaml
stages:
  pre-push:
    steps:
      - action: version-check
      - action: version-greatest
      - action: git-untracked
      - action: git-uncommitted
      - action: git-modified
```

**Current Behavior:**
- If `git-untracked` fails, `git-uncommitted` and `git-modified` never run
- User only sees the first failure, missing other potential issues
- Output order may not match the YAML declaration order

**Desired Behavior:**
- All steps should run regardless of individual failures
- Output should show steps in YAML declaration order
- User gets complete picture of all check results

### 3. Requirements

#### 3.1 Functional Requirements

**FR-1: Ordered Output Display**
- Steps must be displayed in the exact order they appear in `project.yml`
- Output order must be consistent regardless of execution order
- Step numbering should reflect declaration order, not execution order

**FR-2: Parallel Execution on Failure**
- When a step fails, continue executing all other steps that can run independently
- Steps that depend on failed steps should be marked as "SKIPPED" with clear reason
- All step results must be collected and displayed in the final summary

**FR-3: Enhanced Error Reporting**
- Failed steps should clearly indicate their failure status
- Skipped steps should show why they were skipped (dependency on failed step)
- Summary should show counts for OK, WARN, ERROR, and SKIPPED statuses

**FR-4: Dependency-Aware Execution**
- Steps with `require` dependencies on failed steps should be skipped
- Steps without dependencies on failed steps should continue running
- DAG execution should respect both success and failure states

#### 3.2 Non-Functional Requirements

**NFR-1: Performance**
- Parallel execution improvements should not significantly impact performance
- Memory usage should remain reasonable even with many parallel steps
- Execution time should be optimized for independent step execution

**NFR-2: Reliability**
- All step results must be captured and displayed
- No step results should be lost due to early termination
- Error handling must be robust for all execution scenarios

**NFR-3: User Experience**
- Output must be clear and easy to understand
- Status indicators must be consistent and intuitive
- Error messages must provide actionable information

### 4. Technical Specifications

#### 4.1 Output Order Management

**Implementation Approach:**
- Maintain a separate "display order" array based on YAML declaration order
- Execute steps in DAG topological order for optimal performance
- Display results in declaration order for user consistency

**Data Structure:**
```go
type StepDisplay struct {
    StepName    string
    DisplayOrder int
    Status      StepStatus
    Result      StepResult
}
```

#### 4.2 Parallel Execution Engine

**Current Architecture:**
```
DAG Builder → Topological Sort → Sequential Wave Execution → Early Termination on Failure
```

**New Architecture:**
```
DAG Builder → Topological Sort → Parallel Wave Execution → Continue on Failure → Collect All Results
```

**Key Changes:**
- Remove early termination on step failure
- Implement result collection for all steps
- Add dependency checking for step execution
- Maintain execution order separate from display order

#### 4.3 Status Management

**New Status Types:**
- `OK`: Step completed successfully
- `WARN`: Step completed with warnings (onerror: warn)
- `ERROR`: Step failed (onerror: stop)
- `SKIPPED`: Step skipped due to dependency on failed step

**Status Transitions:**
```
PENDING → RUNNING → OK/WARN/ERROR
PENDING → SKIPPED (if dependency failed)
```

### 5. Implementation Plan

#### Phase 1: Output Order Management
1. Modify step display logic to use declaration order
2. Update UI rendering to show steps in correct order
3. Add step numbering based on declaration order

#### Phase 2: Parallel Execution Improvements
1. Remove early termination on step failure
2. Implement dependency-aware step execution
3. Add result collection for all steps
4. Implement SKIPPED status for dependent steps

#### Phase 3: Enhanced Reporting
1. Update summary to include SKIPPED count
2. Improve error messages for skipped steps
3. Add clear indication of why steps were skipped

#### Phase 4: Testing and Validation
1. Create comprehensive test scenarios
2. Test with various dependency configurations
3. Validate output order consistency
4. Performance testing with large step counts

### 6. Success Criteria

#### 6.1 Functional Success
- [ ] Steps display in YAML declaration order
- [ ] All independent steps run regardless of failures
- [ ] Dependent steps are properly skipped with clear messaging
- [ ] Summary shows accurate counts for all status types

#### 6.2 User Experience Success
- [ ] Users can see complete picture of all check results
- [ ] Output is predictable and consistent
- [ ] Error messages are clear and actionable
- [ ] Performance remains acceptable

#### 6.3 Technical Success
- [ ] Code is maintainable and well-tested
- [ ] No regression in existing functionality
- [ ] Performance impact is minimal
- [ ] Error handling is robust

### 7. Risks and Mitigation

#### Risk 1: Performance Impact
- **Risk**: Parallel execution of all steps might impact performance
- **Mitigation**: Implement step execution limits and resource management

#### Risk 2: Memory Usage
- **Risk**: Storing all step results might increase memory usage
- **Mitigation**: Implement efficient result storage and cleanup

#### Risk 3: Complex Dependency Logic
- **Risk**: Dependency checking might become complex
- **Mitigation**: Thorough testing and clear documentation

### 8. Future Considerations

#### 8.1 Potential Enhancements
- Step execution timeouts
- Resource usage monitoring
- Step execution priorities
- Conditional step execution based on previous results

#### 8.2 Backward Compatibility
- All existing configurations must continue to work
- No breaking changes to CLI interface
- Maintain existing error handling behavior where appropriate

### 9. Acceptance Criteria

#### 9.1 Test Scenarios

**Scenario 1: Independent Steps**
```yaml
stages:
  pre-push:
    steps:
      - action: step1
      - action: step2
      - action: step3
```
- All steps should run regardless of individual failures
- Output should show steps in declaration order

**Scenario 2: Dependent Steps**
```yaml
stages:
  pre-push:
    steps:
      - action: step1
      - action: step2
        require: [step1]
      - action: step3
        require: [step1]
```
- If step1 fails, step2 and step3 should be skipped
- If step1 succeeds, step2 and step3 should run in parallel

**Scenario 3: Mixed Dependencies**
```yaml
stages:
  pre-push:
    steps:
      - action: step1
      - action: step2
      - action: step3
        require: [step1]
      - action: step4
        require: [step2]
```
- If step1 fails, step3 should be skipped, step4 should run
- If step2 fails, step4 should be skipped, step3 should run
- If both fail, both step3 and step4 should be skipped

#### 9.2 Performance Requirements
- Execution time should not increase by more than 10%
- Memory usage should not increase by more than 20%
- All existing tests must pass

### 10. Conclusion

This PRD addresses critical user experience issues with the pre-push tool's parallel execution engine. The proposed improvements will provide users with:

1. **Predictable Output**: Steps displayed in declaration order
2. **Complete Information**: All step results regardless of failures
3. **Better Debugging**: Clear indication of why steps were skipped
4. **Improved Reliability**: Robust error handling and result collection

The implementation plan provides a clear path forward with minimal risk and maximum benefit to users.