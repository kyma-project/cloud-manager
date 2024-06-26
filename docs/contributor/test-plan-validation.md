# Validation Test Plan

## Purpose

The validation phase focuses on functionally testing the features requested by customers to ensure they work correctly and are complete. This involves eliciting the requirements and defining them in a human-readable format. Later, this format is used to create an automated acceptance test suite, making the validation process an automated step in the CI/CD pipeline.

## Scope

The validation test suite is the highest-level test suite, which tests the fully integrated system end-to-end to ensure it conforms to user requirements.

## Definitions

Functional testing — a type of software testing where the basic functionalities of an application are tested against a predetermined set of specifications.

Gherkin — non-technical and human readable language for writing software requirements.

Black box testing — a form of testing that is performed with no knowledge of a system's internals.

SDLC (software development life cycle) — is a process for planning, creating, testing, and deploying an information system.

## Audience

The key stakeholder for the validation testing is the `Product Owner` who elicits and formalises the requirements in the `Gherkin` format.

`Software engineers` implement a test suite to automatically test the fully integrated system against the acceptance criteria defined by the product owner.

## Context of testing

### Objective

The objective of validation testing is to ensure that the system provides all functionality requested by user in a consistend and predictable manner.

### Test scope

The system is tested end-to-end in a full integration as a black-box.

<THE DIAGRAM OF THE CLOUD MANAGER COMPONENT IN THE KYMA LANDSCAPE>

### Requirements

The software requirements capturing is a part of the SDLC and should happen in the very beginning of the new feature development.

<SDLC DIAGRAM FOR REQUIREMENTS DEFINITION AND VALIDATION>

The software requirements format is a [Gherkin syntax](https://cucumber.io/docs/gherkin/reference/). 

The software requirements location is the issue-tracking system.

### Test deliverables

The validation test suite provides a report with a listing of all defined user requirements and a current status of checking the systen's conformity to them.

### Environment description

Descripton goes here...

### Entry and exit criteria to each phase

The entry criteria for the testing phase is the list of software requirements in the Gherkin format that are collected from the repository.

The exit criteria is the validation report with a status for each test step in the software requirements.

### Appendix A. Quality evidence

* Link to the filter for the features without requirements.
* Link to the validation report.
* Link to the DoD for writing the requirement.

### Appendix B. Technical gap remediation

<The technical gap identification and the remediation plan for addressing it>