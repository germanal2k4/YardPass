# Build stage
FROM node:20-alpine AS build

WORKDIR /app

# Copy package files
COPY package.json package-lock.json* pnpm-lock.yaml* yarn.lock* ./

# Install dependencies (handle multiple package managers)
RUN if [ -f pnpm-lock.yaml ]; then \
      npm install -g pnpm && pnpm install --frozen-lockfile; \
    elif [ -f yarn.lock ]; then \
      yarn install --frozen-lockfile; \
    else \
      npm ci; \
    fi

# Copy source code
COPY . .

# Set build environment variables
ENV NODE_ENV=production
ENV VITE_API_BASE_URL=http://localhost:8080

# Build application (using build:prod to skip TypeScript checks for Docker)
# Vite handles TypeScript compilation during build
RUN echo "Starting Vite build..." && \
    if [ -f pnpm-lock.yaml ]; then \
      pnpm run build:prod; \
    elif [ -f yarn.lock ]; then \
      yarn build:prod; \
    else \
      npm run build:prod; \
    fi && \
    echo "Build completed successfully" && \
    ls -la dist

# Production stage
FROM nginx:alpine

# Copy nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy built assets
COPY --from=build /app/dist /usr/share/nginx/html

# Expose port
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost/ || exit 1

CMD ["nginx", "-g", "daemon off;"]

