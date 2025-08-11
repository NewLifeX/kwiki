---
title: Installation
type: guide
order: 3
---

# Installation Guide Generation Prompt (English)

Generate a detailed Installation guide for the following project:

**Project Information:**
- Project: {{.ProjectName}}
- Primary Language: {{.PrimaryLanguage}}
- Description: {{.Description}}

**Requirements:**
Create a comprehensive installation guide that includes:

## 1. System Requirements
- **Operating Systems**
  - Supported OS versions
  - Architecture requirements (x86, ARM, etc.)
  - Minimum system specifications
- **Software Dependencies**
  - Required runtime environments
  - Version compatibility matrix
  - Package managers needed
- **Hardware Requirements**
  - Minimum RAM and storage
  - Network connectivity needs
  - Special hardware considerations

## 2. Pre-installation Setup
- **Environment Preparation**
  - System updates and patches
  - Security considerations
  - User permissions and access rights
- **Dependency Installation**
  - Step-by-step dependency setup
  - Version verification commands
  - Troubleshooting dependency issues

## 3. Installation Methods
- **Package Manager Installation** (Recommended)
  - Official package repositories
  - Third-party package managers
  - Automated installation scripts
- **Binary Installation**
  - Pre-compiled binary downloads
  - Installation from archives
  - Path configuration
- **Source Installation**
  - Source code compilation
  - Build tools and requirements
  - Custom build options
- **Container Installation**
  - Docker installation
  - Container orchestration
  - Volume and network configuration

## 4. Step-by-Step Installation
For each installation method, provide:
- **Detailed Commands**
  - Copy-pasteable command sequences
  - Platform-specific variations
  - Expected output examples
- **Configuration Steps**
  - Initial configuration files
  - Environment variable setup
  - Service registration (if applicable)
- **Verification Procedures**
  - Installation success checks
  - Functionality tests
  - Performance validation

## 5. Post-Installation Configuration
- **Initial Setup**
  - First-time configuration wizard
  - Default settings explanation
  - Security hardening steps
- **Integration Setup**
  - System service integration
  - Database connections
  - External service configuration
- **User Account Setup**
  - Admin account creation
  - User permission configuration
  - Authentication setup

## 6. Verification and Testing
- **Installation Verification**
  - Version check commands
  - Service status verification
  - Connectivity tests
- **Functionality Testing**
  - Basic operation tests
  - Feature validation
  - Performance benchmarks
- **Health Checks**
  - System resource usage
  - Log file verification
  - Error detection

## 7. Troubleshooting
- **Common Installation Issues**
  - Permission problems
  - Dependency conflicts
  - Network connectivity issues
- **Platform-Specific Problems**
  - Windows-specific issues
  - macOS-specific issues
  - Linux distribution differences
- **Resolution Steps**
  - Diagnostic commands
  - Log analysis
  - Recovery procedures

## 8. Uninstallation
- **Clean Removal Process**
  - Uninstallation commands
  - Configuration cleanup
  - Data backup procedures
- **Rollback Procedures**
  - Version downgrade steps
  - Configuration restoration
  - Data recovery

**Format Requirements:**
- Use clear, numbered step-by-step instructions
- Include code blocks with syntax highlighting
- Provide platform-specific sections where needed
- Include verification commands for each step
- Add troubleshooting sections for common issues
- Use consistent formatting throughout

**Style Guidelines:**
- Write for users with varying technical expertise
- Provide both quick and detailed installation paths
- Include safety warnings for destructive operations
- Use encouraging, supportive language
- Provide context for each step
