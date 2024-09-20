# build 
FROM node:14

WORKDIR /app

# COPY package*.json ./
COPY . .
RUN npm install
RUN npm run build

# run
FROM nginx:alpine
COPY --from=0 /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 3000

CMD ["nginx", "-g", "daemon off;"]