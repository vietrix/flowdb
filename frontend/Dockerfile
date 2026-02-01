# frontend/Dockerfile
FROM node:20-alpine AS build

WORKDIR /app

COPY package.json pnpm-lock.yaml ./

RUN corepack enable && pnpm install --frozen-lockfile

COPY . .

ARG NEXT_PUBLIC_API_BASE
ENV NEXT_PUBLIC_API_BASE=${NEXT_PUBLIC_API_BASE}
ENV NODE_ENV=production

RUN pnpm build

FROM node:20-alpine AS runtime

WORKDIR /app

ENV NODE_ENV=production
ENV PORT=3000

RUN addgroup -S app && adduser -S -G app app

COPY --from=build /app/package.json /app/pnpm-lock.yaml /app/next.config.mjs /app/tsconfig.json /app/
COPY --from=build /app/public /app/public
COPY --from=build /app/.next /app/.next
COPY --from=build /app/node_modules /app/node_modules

RUN chown -R app:app /app

USER app

EXPOSE 3000

CMD ["node_modules/.bin/next", "start", "-p", "3000"]
