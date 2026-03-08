# Infrastructure Validator

Autonomous infrastructure validator for go-agent development environment.

## Features

- **Docker Container Scanner**: Lists and validates running containers
- **Qdrant Collection Validator**: Checks vector dimensions and auto-creates missing collections
- **API Connectivity Tester**: Tests DeepSeek, OpenAI, Google, Anthropic, Ollama APIs
- **Environment Validator**: Validates and auto-fixes environment variable issues
- **Auto-Fix**: Automatically fixes common issues like dimension mismatches, missing collections

## Quick Start

```bash
# Run once (useful for session startup)
./validate.sh --once

# Run with monitoring loop (10-minute intervals)
./validate.sh --monitor

# Run with verbose output
./validate.sh -v

# Specify expected vector dimension
./validate.sh --dim 1536
```

## Command-Line Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--once`, `-o` | false | Run validation once and exit |
| `--monitor`, `-m` | false | Run with 10-minute monitoring loop |
| `--verbose`, `-v` | false | Enable verbose output |
| `--dim`, `-d N` | 768 | Expected vector dimension |
| `--json` | false | Output in JSON format |
| `--report-dir` | "" | Directory to save validation reports |
| `--auto-fix` | true | Automatically fix detected issues |

## Direct Go Usage

```bash
# Build and run
go build -o validator ./cmd/validator/
./validator -once -expected-dim 768

# Or use go run
go run ./cmd/validator/ -once -v

# With monitoring
go run ./cmd/validator/ -interval 10m
```

## Validation Categories

### 1. Docker Containers

- Checks Docker daemon connectivity
- Lists all running containers
- Validates container states
- Checks for expected containers (Qdrant, etc.)
- Auto-restarts stopped containers (if `--auto-fix`)

### 2. Qdrant Vector Database

- Tests connectivity to Qdrant
- Lists all collections
- Validates vector dimensions against expected
- Reports compatible embedding models for current dimension
- Auto-creates missing expected collections

### 3. API Connectivity

Tests connectivity to:
- OpenAI (`OPENAI_API_KEY`)
- DeepSeek (`DEEPSEEK_API_KEY`)
- Google/Gemini (`GOOGLE_API_KEY` or `GEMINI_API_KEY`)
- Anthropic (`ANTHROPIC_API_KEY`)
- Ollama (`OLLAMA_HOST`)
- Voyage AI (`VOYAGE_API_KEY`)

Reports available models for each connected provider.

### 4. Environment Variables

- Validates API key configuration
- Checks embedding provider settings (`ADK_EMBED_PROVIDER`, `ADK_EMBED_MODEL`)
- Validates cache configuration (`AGENT_LLM_CACHE_*`)
- Detects proxy settings
- Reports environment variable conflicts
- Auto-detects embedding provider from available API keys

## Output Format

### Human-Readable

```
╔══════════════════════════════════════════════════════════════╗
║  Infrastructure Validation Report                              ║
║  2026-03-02 15:04:05 | Duration: 1.2s                         ║
╠══════════════════════════════════════════════════════════════╣
║  Overall: ✅ healthy                                           ║
╠══════════════════════════════════════════════════════════════╣
║  Summary: 12 checks | 10 passed | 2 warnings | 0 errors | 0 fixed  ║
╠══════════════════════════════════════════════════════════════╣
║  🐳 Docker Containers                                         ║
║    ✅ Docker Daemon: Connected (2ms)                          ║
║    ✅ Expected: qdrant: Running on image qdrant/qdrant        ║
║  🗄️  Qdrant Vector Database                                   ║
║    ✅ Qdrant Connectivity: Connected (5ms)                    ║
║    ✅ Collection adk_memories (dim=768): Dimensions match     ║
║  🔌 API Connectivity                                          ║
║    ✅ OpenAI: Connected (150ms, 10 models available)          ║
║    ⚠️  DeepSeek: API key not configured (DEEPSEEK_API_KEY)    ║
║  🔧 Environment Variables                                     ║
║    ✅ API Key: OpenAI: Configured (OPENAI_API_KEY=sk-****)    ║
╚══════════════════════════════════════════════════════════════╝
```

### JSON Format (with `--json`)

```json
{
  "timestamp": "2026-03-02T15:04:05.123Z",
  "duration": 1234567890,
  "overall_status": "healthy",
  "summary": {
    "total_checks": 12,
    "passed": 10,
    "warnings": 2,
    "errors": 0,
    "fixed": 0
  },
  "docker": [...],
  "qdrant": [...],
  "apis": [...],
  "environment": [...]
}
```

## Vector Dimension Reference

| Dimension | Compatible Models |
|-----------|-------------------|
| 384 | sentence-transformers/all-MiniLM-L6-v2 |
| 768 | DeepSeek, Google text-embedding-004, Ollama nomic-embed-text |
| 1024 | Voyage voyage-2, Ollama mxbai-embed-large |
| 1536 | OpenAI text-embedding-ada-002, text-embedding-3-small |
| 3072 | OpenAI text-embedding-3-large |

## Integration

### Shell Profile

Add to `~/.zshrc` or `~/.bashrc`:

```bash
# Run validator at session start
alias validate='~/MareCogito/go-agent/cmd/validator/validate.sh --once'
alias validate-monitor='~/MareCogito/go-agent/cmd/validator/validate.sh --monitor'
```

### CI/CD

```yaml
# GitHub Actions
- name: Validate Infrastructure
  run: cd go-agent && go run ./cmd/validator/ -once -json > validation-report.json
  
- name: Upload Report
  uses: actions/upload-artifact@v3
  with:
    name: validation-report
    path: validation-report.json
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit
go run ./cmd/validator/ -once || exit 1
```