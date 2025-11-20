FROM python:3.11-slim

# Install basic tools
RUN apt-get update && apt-get install -y \
    git \
    curl \
    vim \
    nodejs \
    npm \
    jq \
    && rm -rf /var/lib/apt/lists/*

# Install real Codex CLI
RUN npm install -g @openai/codex

# Create .codex directory
RUN mkdir -p /root/.codex

# Set working directory
WORKDIR /workspace

# Default command (keep container running)
CMD ["tail", "-f", "/dev/null"]
