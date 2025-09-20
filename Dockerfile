FROM golang:1.25 as build
WORKDIR /src
COPY . .
RUN go mod tidy && CGO_ENABLED=0 go build -o /out/bankapp ./cmd/bankapp

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/bankapp /usr/local/bin/bankapp
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/bankapp"]
