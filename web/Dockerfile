# build
FROM node:18-alpine AS builder

WORKDIR /app

COPY . .
RUN npm install

# Google OAuth Client ID - docker-compose build args로 주입
ARG REACT_APP_GOOGLE_CLIENT_ID
ENV REACT_APP_GOOGLE_CLIENT_ID=$REACT_APP_GOOGLE_CLIENT_ID

# Backend API base URL - docker-compose build args로 주입
ARG REACT_APP_BASE_URL=""
ENV REACT_APP_BASE_URL=$REACT_APP_BASE_URL

RUN npm run build

# run
FROM nginx:alpine
COPY --from=builder /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 3000

CMD ["nginx", "-g", "daemon off;"]
