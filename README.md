# Hackstack CLI

_Stop fretting, start building!_

---

> [!NOTE]
> This CLI is still a beta prototype.

## Overview

When it comes to starting a new project, I often find the initial setup to be the most tedious task.
In addition, I tend to follow the same structures in my Go projects anyway.

The Hackstack CLI is designed to simplify this process by providing a streamlined way to scaffold new
projects using predefined templates. Whether you're a seasoned developer or just getting started, the
Hackstack CLI offers an interactive experience that guides you through the project creation process,
allowing you to focus on what matters most: building your application.

### Features

- Interactive TUI: The CLI includes a user-friendly terminal interface that prompts you for essential
  project information such as name, author, and username.
- Templating Engine: The CLI uses a powerful templating engine to generate project files based on
  your input, ensuring a consistent structure across all your projects.
- Cross-Platform: The CLI is designed to work seamlessly across different operating systems, ensuring
  that you can use it regardless of your development environment.

### Stack

- Go 1.24
- Cobra for command-line interface
- BubbleTea for interactive prompts
- Logrus for logging

## Usage

```bash
hackstack build <category> [--source <path-to-yaml>] [--force]
```

## Testing

### Unit tests

```bash
# Run standard assertions with go-test
just test
```

### Automation

#### GitHub Actions integration

Tests are automatically run on:

- Every pull request
- Every push to main branch

#### Quality gates

- All tests must pass before merging
- Minimum code coverage requirements
- Performance benchmarks must be met

## License

This project is licensed under the BSD-3 License. See the LICENSE file for more details.
