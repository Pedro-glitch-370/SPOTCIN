# ---------- Stage 1: build ----------
FROM node:22-alpine AS builder

WORKDIR /app

# Copia só os manifests primeiro
COPY package.json package-lock.json* ./

RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi

# Copia o restante do código
COPY . .

# Build de produção
RUN npm run build

# ---------- Stage 2: runtime com nginx ----------
FROM nginx:1.27-alpine

# Config customizada do nginx
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copia os arquivos estáticos gerados pelo Vite
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]