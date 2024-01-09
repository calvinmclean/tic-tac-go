FROM golang:1.20-alpine AS build
RUN mkdir /build
ADD . /build
WORKDIR /build
RUN go build -o tic-tac-go .

FROM alpine:latest AS production
RUN mkdir /app
WORKDIR /app
COPY --from=build /build/tic-tac-go .
COPY --from=build /build/tictactoe.html .
ENTRYPOINT ["/app/tic-tac-go"]
