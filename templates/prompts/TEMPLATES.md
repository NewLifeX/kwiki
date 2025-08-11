# KWiki Prompt Templates

This directory contains prompt templates for generating different types of documentation in multiple languages.

## Directory Structure

```
templates/prompts/
├── config.yaml          # Template configuration
├── TEMPLATES.md         # This file
├── en/                  # English templates
│   ├── readme.md
│   ├── getting-started.md
│   ├── api-reference.md
│   └── ...
├── zh/                  # Chinese templates
│   ├── readme.md
│   ├── getting-started.md
│   └── ...
└── [other languages]/
```

## Available Template Types

- **readme**: Project overview and main documentation
- **getting-started**: Quick start guide for new users
- **installation**: Detailed installation instructions
- **architecture**: System architecture and design documentation
- **api-reference**: API documentation and reference
- **examples**: Code examples and tutorials
- **configuration**: Configuration guide and options
- **troubleshooting**: Common issues and solutions

## Supported Languages

- `en` - English
- `zh` - 中文 (Chinese)
- `ja` - 日本語 (Japanese)
- `ko` - 한국어 (Korean)
- `es` - Español (Spanish)
- `fr` - Français (French)
- `de` - Deutsch (German)
- `ru` - Русский (Russian)
- `pt` - Português (Portuguese)
- `it` - Italiano (Italian)

## Template Variables

All templates have access to the following variables:

### Project Information
- `{{.ProjectName}}` - Repository name
- `{{.Description}}` - Repository description
- `{{.PrimaryLanguage}}` - Primary programming language
- `{{.License}}` - License information
- `{{.Language}}` - Target documentation language

### Structure Information
- `{{.Modules}}` - Array of modules with the following fields:
  - `{{.Name}}` - Module name
  - `{{.Description}}` - Module description
  - `{{.Functions}}` - Array of functions with:
    - `{{.Name}}` - Function name
    - `{{.Description}}` - Function description

## Template Syntax

Templates use Go's `text/template` syntax:

### Basic Variables
```
Project: {{.ProjectName}}
Description: {{.Description}}
```

### Conditionals
```
{{if .License}}
License: {{.License}}
{{end}}
```

### Loops
```
## Modules
{{range .Modules}}
- **{{.Name}}**: {{.Description}}
{{end}}
```

## Customizing Templates

### Using CLI
```bash
# List available templates
kwiki template list en

# Copy a template for customization
kwiki template copy en readme my-custom-readme.md

# Validate your template
kwiki template validate my-custom-readme.md

# Test your template
kwiki template test en readme
```

### Direct File Editing
1. Copy the template file you want to customize
2. Edit using any text editor
3. Place in the appropriate language directory
4. KWiki will automatically use your custom template

## Template Guidelines

### Content Structure
- Start with clear section headings
- Use consistent formatting
- Include specific requirements for each section
- Provide context about the target audience

### Language Considerations
- Use appropriate technical terminology for each language
- Maintain cultural sensitivity
- Consider local documentation conventions
- Ensure translations are accurate and natural

### AI Prompt Best Practices
- Be specific about desired output format
- Include clear structure requirements
- Specify technical depth required
- Provide examples of desired output
- Include any constraints or limitations

## Examples

### Basic Template Structure
```markdown
# Document Type Generation Prompt (Language)

Generate [document type] for the following project:

**Project Information:**
- Project: {{.ProjectName}}
- Description: {{.Description}}
- Language: {{.PrimaryLanguage}}

**Requirements:**
Create documentation that includes:

## 1. Section One
- Requirement details
- Format specifications

**Format Requirements:**
- Use Markdown format
- Include code examples
- Maintain professional tone
```

### Advanced Template with Loops
```markdown
{{range .Modules}}
### {{.Name}}
{{.Description}}

#### Functions
{{range .Functions}}
- **{{.Name}}**: {{.Description}}
{{end}}
{{end}}
```

## Contributing

When contributing new templates or languages:

1. Follow the existing directory structure
2. Use consistent variable names
3. Include comprehensive documentation requirements
4. Test templates with sample data
5. Ensure translations are accurate and culturally appropriate
