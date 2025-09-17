# Contributing to [Project Name]

First off, thank you for considering contributing to [Project Name]! It's people like you that make open source such a great community. We're excited you're here.

Following these guidelines helps to communicate that you respect the time of the developers managing and developing this open-source project. In return, they should reciprocate that respect in addressing your issue, assessing changes, and helping you finalize your pull requests.

Please note that this project is released with a Contributor [Code of Conduct](CODE_OF_CONDUCT.md). By participating in this project you agree to abide by its terms.

## Table of Contents
- [How Can I Contribute?](#how-can-i-contribute)
  - [Reporting Bugs](#reporting-bugs)
  - [Suggesting Enhancements](#suggesting-enhancements)
  - [Your First Code Contribution](#your-first-code-contribution)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Local Development Setup](#local-development-setup)
- [Pull Request Process](#pull-request-process)
- [Style Guides](#style-guides)
  - [Git Commit Messages](#git-commit-messages)
  - [Go Code Style](#go-code-style)
- [Questions?](#questions)

## How Can I Contribute?

### Reporting Bugs
Bugs are tracked as [GitHub issues](link/to/your/issues). Before opening a new issue, please perform a quick search to see if the problem has already been reported. If it has, please add a comment to the existing issue instead of opening a new one.

When you are creating a bug report, please include as many details as possible. Fill out the required template, the information it asks for helps us resolve issues faster.

### Suggesting Enhancements
If you have an idea for a new feature or an improvement, please open an issue to discuss it. This allows us to coordinate our efforts and prevent duplication of work. We're always open to new ideas!

### Your First Code Contribution
Unsure where to begin contributing to [Project Name]? You can start by looking through these `good first issue` and `help wanted` issues:
* **Good first issues** - issues which should only require a few lines of code, and a test or two.
* **Help wanted issues** - issues which should be a bit more involved than `good first issue` issues.

## Getting Started

### Prerequisites
- Go 1.22 or later
- Git

### Local Development Setup
1.  **Fork** the repository on GitHub.
2.  **Clone** your fork to your local machine:
    ```bash
    git clone [https://github.com/your-username/](https://github.com/your-username/)[Project-Name].git
    cd [Project-Name]
    ```
3.  **Add the `upstream` remote** to keep your fork in sync with the main repository:
    ```bash
    git remote add upstream [https://github.com/original-owner/](https://github.com/original-owner/)[Project-Name].git
    ```
4.  **Install dependencies**:
    ```bash
    go mod tidy
    ```
5.  **Run the tests** to ensure everything is set up correctly:
    ```bash
    go test -v ./...
    ```

## Pull Request Process
1.  Create a new branch from `main` for your changes. Please give it a descriptive name (e.g., `feat/add-new-button` or `fix/login-bug`).
    ```bash
    git checkout -b feat/my-awesome-feature
    ```
2.  Make your changes, ensuring you follow the style guides.
3.  Add or update tests for your changes. We require good test coverage.
4.  Ensure all tests pass before submitting.
5.  Commit your changes using a descriptive commit message that follows our [commit message guidelines](#git-commit-messages).
6.  Push your branch to your fork on GitHub:
    ```bash
    git push origin feat/my-awesome-feature
    ```
7.  Open a pull request to the `main` branch of the original repository.
8.  Fill out the pull request template with the required information. Link the PR to any relevant issues.
9.  Your PR will be reviewed by a maintainer. They may request changes. Please be responsive to feedback.

## Style Guides

### Git Commit Messages
We follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification. This allows for automated changelog generation and makes the project history easier to read.

Your commit messages should be structured as follows: