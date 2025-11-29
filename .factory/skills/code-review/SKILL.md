# Code Review Skill

Automated code review for pull requests using multiple specialized agents with confidence-based scoring to filter false positives.

## Overview

This skill automates pull request review by launching multiple agents in parallel to independently audit changes from different perspectives. It uses confidence scoring to filter out false positives, ensuring only high-quality, actionable feedback is posted.

## Usage

When reviewing a pull request, follow these steps precisely:

### Step 1: Eligibility Check
Check if the pull request:
- (a) is closed
- (b) is a draft
- (c) does not need a code review (automated PR, very simple and obviously ok)
- (d) already has a code review

If any of the above, do not proceed.

### Step 2: Gather Guidelines
Find relevant guideline files from the codebase:
- Root `AGENTS.md` or `CLAUDE.md` file (if exists)
- Any guideline files in directories whose files the PR modified

### Step 3: Summarize Changes
View the pull request and create a summary of the changes.

### Step 4: Multi-Agent Review
Launch 5 parallel review passes to independently audit the change:

**Agent #1 - Guideline Compliance**: Audit the changes to ensure they comply with AGENTS.md/CLAUDE.md. Note that these files contain guidance for AI as it writes code, so not all instructions will be applicable during code review.

**Agent #2 - Bug Scan**: Read the file changes in the PR, then do a shallow scan for obvious bugs. Avoid reading extra context beyond the changes, focusing just on the changes themselves. Focus on large bugs, avoid small issues and nitpicks. Ignore likely false positives.

**Agent #3 - Historical Context**: Read the git blame and history of the code modified, to identify any bugs in light of that historical context.

**Agent #4 - Previous PR Comments**: Read previous pull requests that touched these files, and check for any comments on those PRs that may also apply to the current PR.

**Agent #5 - Code Comments**: Read code comments in the modified files, and make sure the changes comply with any guidance in the comments.

### Step 5: Confidence Scoring
For each issue found, score on a scale from 0-100:

| Score | Confidence Level | Description |
|-------|-----------------|-------------|
| 0 | Not confident | False positive that doesn't stand up to light scrutiny, or is a pre-existing issue |
| 25 | Somewhat confident | Might be a real issue, but may be false positive. Unable to verify. Stylistic issues not explicitly in guidelines |
| 50 | Moderately confident | Verified real issue, but might be a nitpick or not happen often. Not very important relative to rest of PR |
| 75 | Highly confident | Double checked and verified very likely real issue that will be hit in practice. Existing approach is insufficient. Very important or directly mentioned in guidelines |
| 100 | Absolutely certain | Confirmed definitely a real issue that will happen frequently. Evidence directly confirms this |

### Step 6: Filter Issues
Filter out any issues with a score less than 80. If no issues meet this criteria, do not proceed with commenting.

### Step 7: Re-verify Eligibility
Repeat the eligibility check from Step 1 to ensure PR is still eligible for review.

### Step 8: Post Review Comment
Comment on the pull request with the result in this format:

```
### Code review

Found N issues:

1. <brief description of bug> (AGENTS.md says "<...>")

<link to file and line with full sha + line range>

2. <brief description of bug> (some/other/AGENTS.md says "<...>")

<link to file and line with full sha + line range>

3. <brief description of bug> (bug due to <file and code snippet>)

<link to file and line with full sha + line range>
```

Or, if no issues found:

```
### Code review

No issues found. Checked for bugs and AGENTS.md compliance.
```

## False Positives to Ignore

These should NOT be flagged:

- **Pre-existing issues** - Problems that existed before this PR
- **Not actually bugs** - Code that looks problematic but isn't
- **Pedantic nitpicks** - Things a senior engineer wouldn't call out
- **Linter/compiler issues** - Missing imports, type errors, formatting (CI will catch these)
- **General code quality** - Lack of tests, general security issues, poor docs (unless explicitly required in guidelines)
- **Silenced issues** - Issues with lint ignore comments
- **Intentional changes** - Functionality changes related to the broader change
- **Unmodified lines** - Real issues on lines the user did not modify

## Link Format

When linking to code, use this exact format:
```
https://github.com/owner/repo/blob/[full-sha]/path/file.ext#L[start]-L[end]
```

Requirements:
- Must use full git SHA (not abbreviated)
- Must use `#L` notation for line numbers
- Must include line range with at least 1 line of context before/after
- Repo name must match the repo being reviewed

## Key Principles

1. **Keep output brief** - Concise descriptions
2. **Avoid emojis** - Professional tone
3. **Link and cite** - Reference relevant code, files, and URLs
4. **Trust the threshold** - 80+ confidence filters most false positives
5. **Don't build/typecheck** - CI handles that separately
6. **Use gh CLI** - For GitHub interactions, not web fetch

## Requirements

- Git repository with GitHub integration
- GitHub CLI (`gh`) installed and authenticated
- AGENTS.md/CLAUDE.md files (optional but recommended)

---

*Source: Adapted from Anthropic's claude-code code-review plugin*
