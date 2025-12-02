# Build Stage
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package.json package-lock.json* ./

# Install dependencies
RUN npm install

# Copy source code
COPY . .

# Build the extension
RUN npm run build

# Production Stage
FROM scratch AS export

# Copy built extension
COPY --from=builder /app/dist /dist

# This stage is used to export the built extension
# Run: docker build --target export --output dist .
