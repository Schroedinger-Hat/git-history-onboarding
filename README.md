# Git History Analyzer

A tool for analyzing Git repositories to generate documentation about feature ownership, development timelines, and code organization.

## Features

- 📊 Repository analysis
- 👥 Feature ownership tracking
- 📝 Feature-to-commit mapping
- 📈 Feature development timelines
- 🐛 Bug tracking and resolution history
- 📚 Automated documentation generation

## Prerequisites

- Go 1.21 or higher
- Git installed on your system
- Access to the Git repositories you want to analyze

## Installation

1. Clone this repository:

```bash
git clone https://github.com/schroedinger-hat/git-history-analyzer.git
```

2. Install dependencies:

```bash
go mod tidy
```

3. Build the project:

```bash
go build -o git-analyzer ./cmd/analyzer
```

## Usage

### Basic Analysis

Analyze a Git repository using the following command:

```bash
./git-analyzer analyze -r <repository-url>
```

### Command Line Options

- `-r, --repo`: Repository URL to analyze (required)

### Example Output

```bash
./git-analyzer analyze https://github.com/schroedinger-hat/git-history-analyzer.git
```

## Project Structure

```bash
git-history-analyzer/
├── cmd/
│ └── analyzer/ # Main application entry point
├── internal/
│ ├── git/ # Git operations (clone, history)
│ ├── analysis/ # Analysis logic
│ │ ├── ownership/ # Code ownership analysis
│ │ ├── features/ # Feature tracking
│ │ └── timeline/ # Story/timeline generation
│ └── models/ # Data structures
└── pkg/ # Public packages if needed
```

### Running Tests

```bash
go test ./...
```

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Future Enhancements

- [ ] Support for multiple branch analysis
- [ ] Integration with issue tracking systems
- [ ] Custom report generation
- [ ] Web interface for visualization
- [ ] Export data in various formats (JSON, CSV, PDF)

## License

This project is licensed under the [Gnu Affero General Public License v3.0](LICENSE)

## Support

For support, please open an issue in the GitHub repository