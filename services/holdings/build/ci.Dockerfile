# This Dockerfile is intended for use as part of a CI pipeline that creates a 
# standalone holdings binary prior to creating a docker image. The holdings
# binary is copied into this scratch image and saves time by preventing us 
# from building the binary twice.
# 
# To build the binary AND the image, use the standard Dockerfile instead.

FROM alpine:latest

RUN apk --no-cache add ca-certificates

LABEL org.opencontainers.image.title="Monay Holdings"
LABEL org.opencontainers.image.source="https://github.com/levisegal/monay"
LABEL org.opencontainers.image.description="Monay Holdings service"

ARG BIN_FILE
COPY ${BIN_FILE} /usr/local/bin/holdings

CMD ["holdings", "server"]



