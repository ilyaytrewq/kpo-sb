# ---------- build ----------
FROM node:20-alpine AS build
WORKDIR /app

COPY services/frontend/package.json ./package.json
COPY services/frontend/vite.config.js ./vite.config.js
COPY services/frontend/index.html ./index.html
COPY services/frontend/src ./src

RUN npm install --no-audit --no-fund
RUN npm run build

# ---------- run -----------
FROM nginx:alpine
COPY services/frontend/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 8080
