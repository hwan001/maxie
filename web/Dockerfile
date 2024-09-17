# Base image
FROM node:14-alpine AS builder

# Set work directory
WORKDIR /app

# Install dependencies
COPY package.json yarn.lock ./
RUN yarn install

# Copy all files
COPY . .

# Build the React app
RUN yarn build

# Final image
FROM nginx:alpine
COPY --from=builder /app/build /usr/share/nginx/html
CMD ["nginx", "-g", "daemon off;"]
