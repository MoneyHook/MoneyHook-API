FROM node:22-bookworm-slim

RUN npm install --global firebase-tools@15.28.2 \
    && npm cache clean --force

WORKDIR /workspace

COPY firebase.json ./firebase.json

EXPOSE 4000 9099

CMD ["firebase", "emulators:start", "--only", "auth", "--project", "demo-moneyhooks"]
