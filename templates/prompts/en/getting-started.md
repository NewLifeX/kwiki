---
title: Getting Started
type: guide
order: 2
---

# Getting Started Guide Generation Prompt (English)

Generate a detailed Getting Started guide for the following project:

**Project Information:**
- Project: {{.ProjectName}}
- Language: {{.PrimaryLanguage}}
- Description: {{.Description}}

**Requirements:**
Create a comprehensive getting started guide that includes:

## 1. Prerequisites
- **System Requirements**
  - Supported operating systems and versions
  - Minimum hardware specifications
  - Required software dependencies
- **Development Environment**
  - Recommended IDEs or editors
  - Essential tools and utilities
  - Environment setup considerations
- **Account Setup** (if applicable)
  - Required accounts or services
  - API keys or credentials
  - Registration processes

## 2. Installation
- **Quick Installation**
  - One-command installation (if available)
  - Package manager commands
  - Pre-built binaries
- **Alternative Methods**
  - Source code compilation
  - Container/Docker setup
  - Cloud deployment options
- **Verification Steps**
  - Commands to verify installation
  - Expected output examples
  - Troubleshooting common issues

## 3. Project Setup
- **Initial Configuration**
  - Configuration file creation
  - Environment variables setup
  - Default settings explanation
- **Project Structure**
  - Directory layout
  - Important files and their purposes
  - Customization options

## 4. Your First Example
- **Hello World Tutorial**
  - Step-by-step walkthrough
  - Complete code example
  - Expected output with screenshots
- **Basic Operations**
  - Common tasks demonstration
  - Interactive examples
  - Variations and modifications

## 5. Next Steps
- **Learning Path**
  - Recommended tutorials
  - Advanced features overview
  - Best practices introduction
- **Resources**
  - Documentation links
  - Community resources
  - Example projects

## 6. Common Issues & Solutions
- **Installation Problems**
  - Permission issues
  - Dependency conflicts
  - Platform-specific problems
- **Runtime Issues**
  - Configuration errors
  - Common error messages
  - Debugging techniques

**Format Requirements:**
- Provide specific, copy-pasteable code examples
- Include expected output for verification steps
- Use clear, numbered instructions
- Add code syntax highlighting
- Include troubleshooting sections
- Provide links to relevant documentation

**Style Guidelines:**
- Write for beginners but don't oversimplify
- Use encouraging, supportive language
- Include practical tips and warnings
- Provide context for each step
