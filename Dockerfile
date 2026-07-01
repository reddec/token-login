FROM scratch
# 8080 for Admin UI and Auth
EXPOSE 8080/tcp
VOLUME /data
ENV DB_URL="sqlite:///data/token-login.sqlite?cache=shared"
ENTRYPOINT ["/token-login"]
ADD token-login /