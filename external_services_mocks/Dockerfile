# syntax=docker/dockerfile:1.7

FROM node:22-alpine
WORKDIR /app

RUN npm install -g json-server@1.0.0-beta.9 \
  && addgroup -S app \
  && adduser -S app -G app

COPY db.json pizza_place_db.json sushi_bar_db.json taco_truck_db.json ./

ENV VENDOR_API_PORT=4010
ENV VENDOR_API_DB_PATH=/app/db.json

EXPOSE 4010

USER app

CMD ["sh", "-c", "json-server --host 0.0.0.0 --watch \"${VENDOR_API_DB_PATH}\" --port \"${VENDOR_API_PORT}\""]
