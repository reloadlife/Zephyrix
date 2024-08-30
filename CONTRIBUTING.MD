# Contributing to Zephyrix

First off, thank you for considering contributing to Zephyrix! It's people like you that make Zephyrix such a great tool.

## Code of Conduct

By participating in this project, you are expected to uphold our [Code of Conduct](CODE_OF_CONDUCT.md).

## How Can I Contribute?

### Reporting Bugs

- Ensure the bug was not already reported by searching on GitHub under [Issues](https://github.com/reloadlife/zephyrix/issues).
- If you're unable to find an open issue addressing the problem, [open a new one](https://github.com/reloadlife/zephyrix/issues/new). Be sure to include a title and clear description, as much relevant information as possible, and a code sample or an executable test case demonstrating the expected behavior that is not occurring.

### Suggesting Enhancements

- First, read the [DEVELOPMENT.md](DEVELOPMENT.md) file to understand our roadmap.
- For new features not listed in DEVELOPMENT.md, start by opening a discussion in the [Discussions](https://github.com/reloadlife/zephyrix/discussions) section. This helps avoid duplicate work and allows the community to weigh in on the proposal.
- Clearly describe the feature and its use case. Provide examples of how it would be used if possible.

### Pull Requests

1. Fork the repo and create your branch from `main`.
2. If you've added code that should be tested, add tests.
3. Ensure your code adheres to the existing style. We use `gofmt` for formatting.
4. Ensure the test suite passes.
5. Make sure your code lints.

## Development Process

1. Create a new branch: `git checkout -b my-branch-name`
2. Make your changes, adding tests as necessary
3. Run the test suite and ensure all tests pass
4. Push to your fork and [submit a pull request](https://github.com/reloadlife/zephyrix/compare)
5. Pat yourself on the back and wait for your pull request to be reviewed and merged.

## Testing

- We strive for the highest possible test coverage. Always write tests for your code.
- Aim for both unit tests and integration tests where applicable.
- Consider using AI tools to help generate comprehensive test cases and increase coverage.
- Run `go test ./... -cover` to check your test coverage.

## Style Guide

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Run `gofmt` on your code before committing.

## Documentation

- Update the README.md with details of changes to the interface, this includes new environment variables, exposed ports, useful file locations, and container parameters.
- Update the DEVELOPMENT.md file if you've completed a feature or added a new one to the roadmap.
- Maintain in-code documentation with clear and concise comments.

<!-- ## Community

- Join our [Discord server](https://discord.gg/zephyrix) for real-time discussion and support.
- Participate in code reviews. Fresh eyes can catch a lot of things that may be missed. -->

## Git Commit Messages

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or less
- Reference issues and pull requests liberally after the first line

## Questions?

Don't hesitate to ask questions in our [Discussions](https://github.com/reloadlife/zephyrix/discussions) section. We're here to help!

Remember, contributions to Zephyrix should be a fun and rewarding experience. If you find any part of the contribution process difficult or frustrating, please let us know so we can improve the process for everyone.

Thank you for contributing to Zephyrix!

<!--  this is also ai generated and will change soon -->
