FROM golang:1.16-alpine AS build

WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -a -o ./live_pulse

FROM scratch

COPY --from=build src/live_pulse .

EXPOSE 8000

ENTRYPOINT [ "./live_pulse" ]