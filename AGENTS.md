# conclave

- AI-powered multi-agent deliberation platform that orchestrates debates and council discussions between AI agents with diverse personas. It is using existing CLI AI agents installed on the system.

- Primary focus on on web UI.

# Commands

- `make serve` - Start the conclave web server. Log is at logs/dev.log

# How to test

- Note: Prioritize Playwright MCP tests for web UI functionality.
- Example:
  - Go to `http://localhost:8182`.
  - Question: "Should I use microservices? 3 engineers total.". Submit

# How to add, update features

- Always add tests for new features.
- Only verify via Playwright MCP once the features are verified via gotest. Playwright are expensive and slow, so only run them when necessary.

# How to run each agent individually

- claude: `claude --output-format json --model sonnet --print "Should I use microservices? In 5 words or less."`
- gemini: `gemini --output-format json --model gemini-3-flash-preview "Should I use microservices? In 5 words or less."`
- opencode: `opencode --format json --model zai-coding-plan/glm-4.7 run "Should I use microservices? In 5 words or less."`
- qwen: `qwen --output-format json --model qwen3-coder-plus-2025-09-23 -p "Should I use microservices? In 5 words or less."`
