# Multi-stage build for Next.js frontend
FROM node:18-alpine AS base

# Install dependencies only when needed
FROM base AS deps
RUN apk add --no-cache libc6-compat
WORKDIR /app

# Install dependencies based on the preferred package manager
COPY package.json package-lock.json* ./
RUN npm ci --ignore-scripts

# Rebuild the source code only when needed
FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

# Accept build arguments for production configuration
ARG NEXT_PUBLIC_API_URL
ARG NEXT_PUBLIC_WS_URL
ARG NODE_ENV
ARG MODE
ARG NEXT_PUBLIC_ENABLE_ANALYTICS
ARG NEXT_PUBLIC_ENABLE_DEBUG
ARG NEXT_PUBLIC_FIUU_SANDBOX
ARG NEXT_PUBLIC_LIVEKIT_URL

# Set environment variables for build
ENV NEXT_TELEMETRY_DISABLED 1
ENV NODE_ENV=${NODE_ENV:-production}
ENV MODE=${MODE:-remote}
ENV NEXT_PUBLIC_API_URL=${NEXT_PUBLIC_API_URL:-http://gateway:8092}
ENV NEXT_PUBLIC_WS_URL=${NEXT_PUBLIC_WS_URL:-ws://gateway:8092}
ENV NEXT_PUBLIC_ENABLE_ANALYTICS=${NEXT_PUBLIC_ENABLE_ANALYTICS:-false}
ENV NEXT_PUBLIC_ENABLE_DEBUG=${NEXT_PUBLIC_ENABLE_DEBUG:-false}
ENV NEXT_PUBLIC_FIUU_SANDBOX=${NEXT_PUBLIC_FIUU_SANDBOX:-true}
ENV NEXT_PUBLIC_LIVEKIT_URL=${NEXT_PUBLIC_LIVEKIT_URL:-wss://blytz-live-u5u72ozx.livekit.cloud}

# Build the application
RUN npm run build

# Production image, copy all the files and run next
FROM base AS runner
WORKDIR /app

# Set production environment variables
ENV NODE_ENV production
ENV NEXT_TELEMETRY_DISABLED 1
ENV PORT 3000
ENV HOSTNAME "0.0.0.0"

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

# Copy public directory if it exists (optional for Next.js)
RUN if [ -d "/app/public" ]; then cp -r /app/public ./public; fi || true

# Set the correct permission for prerender cache
RUN mkdir .next
RUN chown nextjs:nodejs .next

# Automatically leverage output traces to reduce image size
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000

ENV PORT 3000
ENV HOSTNAME "0.0.0.0"

CMD ["node", "server.js"]