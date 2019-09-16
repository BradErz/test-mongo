FROM golang:alpine AS build-env
ADD . /src
RUN cd /src && go build -o test-mongo

# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /src/test-mongo /app/
ENTRYPOINT ./test-mongo