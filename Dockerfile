FROM golang:1.22rc1-bullseye

ENV PATH="$PATH:$GOROOT/bin:$GOPATH/bin"
ENV PATH="$PATH:$(go env GOPATH)/bin"

COPY . .

RUN apt-get update \
	&& apt-get install -y \
	&& go install -v github.com/air-verse/air@v1.52.3 \
	&& go install -v golang.org/x/tools/gopls@v0.11.0 \
	&& go install -v github.com/cweill/gotests/gotests@v1.6.0 \
	&& go install -v github.com/fatih/gomodifytags@v1.16.0 \
	&& go install -v github.com/josharian/impl@v1.1.0 \
	&& go install -v github.com/haya14busa/goplay/cmd/goplay@v1.0.0 \
	&& go install -v github.com/go-delve/delve/cmd/dlv@v1.9.0 \
	&& go install -v honnef.co/go/tools/cmd/staticcheck@v0.4.3 \
	&& go install -v golang.org/x/tools/cmd/goimports@v0.1.10

# WORKDIR /workspace/app
# RUN go mod tidy

ENV CGO_ENABLED=0 \
	GOOS=linux \
	GOARCH=arm64

EXPOSE 8080

# CMD ["go", "run", "main.go"]
# CMD ["air", "-c", ".air.toml"]
# ENTRYPOINT cd app && go run main.go
