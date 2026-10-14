FROM node:22-alpine AS build

WORKDIR /app
ENV NEXT_TELEMETRY_DISABLED=1

COPY package.json package-lock.json ./
COPY client/package.json ./client/package.json
RUN npm ci

COPY client ./client
RUN npm run build --workspace client

FROM node:22-alpine

WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

COPY --from=build /app/package.json /app/package-lock.json ./
COPY --from=build /app/node_modules ./node_modules
COPY --from=build /app/client ./client

EXPOSE 3000
CMD ["npm", "run", "start", "--workspace", "client", "--", "-H", "0.0.0.0", "-p", "3000"]
