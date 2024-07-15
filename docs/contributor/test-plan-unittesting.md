# Unittesting Test Plan

## **Purpose** 

The unit testing phase focuses on maintainability of the code. This involves testing the logic of smallest grains of the code in isolation and it aims to prevent regression resulted by code changes. 

## Scope

Unit test suits are the lowest-level test suites, and they test the smallest grains of logic, mainly methods, in isolation, to prevent regression resulted by code changes. Unit tests should be able to run in isolation independent of external systems such as databases, external APIs, etc. 

When the code is using any such dependencies, they should be mocked by using piece of logic that would be running in the local process space.

## Definitions

Dependency Injection — a technique whereby the dependencies for a logic, is passed to it instead of having a hard dependency in the logic itself.

Code Coverage — a report of percentage measure of the degree to which the source code is executed when all unittests are run.

Test Driven Development (TDD) — in unit test context, is to write the tests as soon as the signature of a method is clarified to test the expected logic. This can be done by the same developer who implements the method or another developer.

## Audience

The primary audience for this document are `Team Leads` and `Software engineers`. 

## References

- github.com/stretchr/testify
- https://github.com/vektra/mockery
- https://github.com/kubernetes-sigs/controller-runtime/tree/main/pkg/client/fake

## Context of testing

Unit testing is part of the build and PRs have a hard dependency on all the unit tests to pass. 

When there are different versions of a code, or different feature flags, that can impact the logic of a given unit, each context has to have its own dedicated unit test or the same unit test needs to be run with all the identified contexts. 

### Objective

The objective is to automatically validate the functionality of small and individual code blocks for correct behavior, and it would be executed each time a build is made. It would ensure that any new changes done in the system is not breaking code logic. 

### Test scope

All the units of the code, in cloud-manager repository, and all its features are in the scope.

### Assumptions and constraints

- All the unit tests must pass or skipped as part of the build. 

- Code coverage must remain in an expected predefined treshold for the PR check to succeed.

- For any given source file, the percentage of the code coverage should not drop as part of the PR. (Good to have)

- To make a code testable, dependency injection as a set of design patterns must be used at all times. Having a hard dependency to a specific object in a code, makes unit testing it in isolation extremely difficult or even impossible. 

- Developers are recommended to follow a TDD approach.

- For assertions, testify

- For mocking, developers have two options:

  - Write their own fake/mock object 
  - Use testify/mock and/or mockery if needed.

- For mocking a http server, e.g. a call to gcp, net.http.httptest to be used

- To test a unit in isolation, which calls k8s APIs in its logic, pass fake k8s client instead. 

  ### Code style and naming convention

  - A test should follow this signature: func Test<Method Name>(t *testing.T). <Method Name> should be replaced by the source code method to be unit tested. 
  - When unit testing a method need more than one logical test, as it has conditional execution path depending on the input, one of the following approaches must be taken:
    - A testify.suite.Suite to be created with as many methods as needed to test all execution pathes. 
    - A series of subtests to be executed with this signature: t.Run("name of the test, func(t *testing.T){...}). 
    - Naming of the methods/subtests for both above options should be in upper camel case and should convey the scenario/path it is testing.

## Test strategy

The unit tests will be run as part of each build process, and they should run in isolation independently of any external dependencies. The cloud manager unit tests would use

* Fake client library to mock the Kubernetes environment
* Mocks to simulate the GCP and AWS APIs. Or any other third-party API.

### Test levels

Unit tests are level 1. 

### Test types

Level 1 tests, i.e. unit tests, must always pass and as they run in isolation, there is not scenario that they can't be executed.

### Test deliverables

Code coverage report that shows the coverage percentage of all the separate source files, preferably in a format that provides a visual representation of the covered code. 

Also an aggregate percentage on the entire source code coverage. 

### Entry and exit criteria to each phase

As long as build does not fail, all the unit tests can be executed. Exit criteria: 

- All the unit tests should complete successfully.
- Average coverage for each package should be more than or equal to 85%
- Preferably it should make sure that coverage percentage does not drop as part of a PR for any given source file.

### Test completion criteria

Unit test phase is complete, when all the unit test suites are completed and the coverage and passed/skipped/failed report is generated.

### Metrics to be collected

Following metrics needs to be collected as part of unit test suite execution:

* Aggregate code coverage for the entire cloud-manager repo.
* Aggregate code coverage for a package unit test suite.
* Code coverage for all source files.
* Aggregate number of all passed, skipped and failed tests for each suite.
* For any given failure, the name and the path of the failed test.

### Test environment requirements

To prevent any false positives, unit tests must get executed in an env with enough resources. The minimum required resources are 2 CPU cores and 8 GB of memory. This amount is less than resources available to Github actions.

## Appendix A. Quality Evidence

- For any given PR (to be replaced in following url), the unit test results can be found in https://github.com/kyma-project/cloud-manager/pull/<PR>/checks.
  - For each build under the checks, the results of the test can be found in "Build and test" step.

## Appendix B. Supplementary References