# ci-status

[![Autorelease](https://github.com/lucasew/ci-status/actions/workflows/autorelease.yml/badge.svg)](https://github.com/lucasew/ci-status/actions/workflows/autorelease.yml)

`ci-status` is a cross-platform CLI tool that wraps command execution and automatically reports status to GitHub (other forges are planned, contributions welcome!). It detects the forge context automatically and reports pending/success/failure states based on command exit codes.

## Installation

Install using [mise](https://mise.jdx.dev):

```sh
mise use github:lucasew/ci-status
```

## Dogfooding

This tool is used to report the status of its own build process.
The [Autorelease workflow](.github/workflows/autorelease.yml) bootstraps `ci-status` and then uses it to report the build status of each cross-compilation target back to GitHub.
