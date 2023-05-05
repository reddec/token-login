FROM scratch
# 8080 for Admin UI, 8081 for Auth
EXPOSE 8080/tcp 8081/tcp
VOLUME /data
ENV DB_URL="sqlite:///data/token-login.sqlite?cache=shared"
ENTRYPOINT ["/token-login"]
ADD token-login /