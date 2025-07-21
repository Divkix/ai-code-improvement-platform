# Use Bun for build and Node for runtime (properly)
FROM oven/bun:1-alpine

# Set working directory
WORKDIR /app

# Copy package files and install dependencies
COPY package.json bun.lock* ./
RUN bun install

# Copy source code
COPY . .

# Build the application
RUN bun run build

# Expose port 3000
EXPOSE 3000

# Set environment variables
ENV HOST=0.0.0.0 PORT=3000

# Start the application using Bun
CMD ["bun", "run", "--bun", "build/index.js"]
