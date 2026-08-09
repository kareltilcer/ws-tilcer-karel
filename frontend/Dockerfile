# Static SPA image for the `karel` frontend: Nginx serves the built bundle and
# NOTHING else. The API is NOT proxied here.
#
# In production the SPA and the Go API share one origin (karel.tilcer.cz): the
# Coolify/Traefik layer routes karel.tilcer.cz/api (and /healthz, /readyz) to the
# backend app, and everything else to this image. The SPA calls the API
# host-relative (relative /api — see src/api/client.ts), so there is no nginx
# proxy and no CORS. The backend app is on port 2002; this image on 2001.
#
# The offline docker-compose harness has no Traefik, so it builds with
# NGINX_CONF=nginx.harness.conf to add a /api proxy to the `backend` service.

# ── Stage 1: build the SPA ─────────────────────────────────────────────
FROM node:22-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
# Vite bakes VITE_* at build time.
ARG VITE_TURNSTILE_SITE_KEY=
ENV VITE_TURNSTILE_SITE_KEY=${VITE_TURNSTILE_SITE_KEY}
RUN npm run build

# ── Stage 2: serve with Nginx ──────────────────────────────────────────
FROM nginx:1.27-alpine
ARG NGINX_CONF=nginx.conf
COPY ${NGINX_CONF} /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 2001
CMD ["nginx", "-g", "daemon off;"]
