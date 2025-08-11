---
title: README
type: overview
order: 1
---

# README Page Generation Prompt (English)

Generate a comprehensive README documentation for the following code repository:

**Project Information:**
- Project Name: {{.ProjectName}}
- Description: {{.Description}}
- Primary Language: {{.PrimaryLanguage}}
- License: {{.License}}

**Project Structure:**
{{range .Modules}}
- {{.Name}}: {{.Description}}
{{end}}

**Requirements:**
Generate a README document that includes:

## 1. Project Overview
- Brief description of what the project does
- Key benefits and use cases
- Target audience
- Project status and maturity level

## 2. Key Features
- List of main features with brief descriptions
- What makes this project unique
- Supported platforms/environments
- Performance characteristics

## 3. Quick Start
- Minimal example to get started immediately
- Basic usage demonstration
- Expected output
- Link to detailed documentation

## 4. Installation
- System requirements
- Installation steps for different platforms
- Package manager commands
- Verification commands

## 5. Basic Usage
- Common use cases with examples
- Code snippets with explanations
- Configuration options
- Command-line interface (if applicable)

## 6. Documentation
- Links to detailed documentation
- API reference
- Tutorials and guides
- Examples repository

## 7. Contributing
- How to contribute to the project
- Development setup instructions
- Coding standards and guidelines
- Issue reporting process

## 8. Support
- Where to get help
- Community resources
- FAQ section
- Contact information

## 9. License
- License information
- Copyright notice
- Third-party licenses

**Format Requirements:**
- Use clear, professional Markdown formatting
- Include code blocks with appropriate syntax highlighting
- Add relevant badges (build status, version, downloads, etc.)
- Use proper heading hierarchy
- Include a table of contents for longer documents
- Ensure all links are functional
- Use consistent formatting throughout

**Style Guidelines:**
- Write in clear, concise English
- Use active voice where possible
- Include practical examples
- Be welcoming to newcomers
- Maintain professional tone
