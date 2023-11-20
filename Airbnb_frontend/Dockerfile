FROM node:18-alpine AS build_container
WORKDIR /usr/src/app
COPY package*.json ./
RUN npm install -g @angular/cli@latest
RUN npm install
COPY . .
RUN npm run build
EXPOSE 4200
CMD ["npm", "start"]
