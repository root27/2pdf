FROM golang:1.22-alpine as builder

WORKDIR /app

COPY . .

RUN go mod download


RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o /bin/2pdf .



FROM alpine

WORKDIR /bin

COPY --from=builder /bin/2pdf /bin/2pdf

COPY --from=builder /app/templates /bin/templates

RUN apk add libreoffice \ 
	build-base \ 
	# Install fonts
	msttcorefonts-installer fontconfig && \
	update-ms-fonts && \
	fc-cache -f

RUN apk add --no-cache build-base libffi libffi-dev



ENTRYPOINT ["/bin/2pdf"]




