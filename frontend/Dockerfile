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

# Set build environment variables (VITE_API_BASE_URL="" for same-origin in k8s)
ARG VITE_API_BASE_URL=http://localhost:8080
ENV NODE_ENV=production
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL

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

# Copy nginx template and entrypoint
COPY nginx.conf.template /etc/nginx/conf.d/default.conf.template
COPY docker-entrypoint.sh /docker-entrypoint-custom.sh
RUN chmod +x /docker-entrypoint-custom.sh

# Copy built assets
COPY --from=build /app/dist /usr/share/nginx/html

ENV BACKEND_URL=http://backend:8080

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost/ || exit 1

ENTRYPOINT ["/docker-entrypoint-custom.sh"]

