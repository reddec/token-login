FROM --platform=$BUILDPLATFORM alpine:3.20 AS certs
RUN apk add --no-cache ca-certificates && update-ca-certificates

FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# 8080 for Admin UI and Auth
EXPOSE 8080/tcp
VOLUME /data
ENV DB_URL="sqlite:///data/token-login.sqlite?cache=shared"
ADD token-login /
ENTRYPOINT ["/token-login"]