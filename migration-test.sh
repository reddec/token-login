#!/usr/bin/env bash
set -euo pipefail

PREV_VERSION="1.1.0"
TARGET_VERSION="snapshot"
# BUILD current version

make snapshot

echo "Starting postgres in background"
POSTGRES=$(docker run -h postgres -e POSTGRES_PASSWORD=postgres -q -P -d "postgres:14")

for DBURL in "sqlite:///data/token-login.sqlite?cache=shared" "postgres://postgres:postgres@postgres:5432/postgres?sslmode=disable"; do
  echo "Testing $DBURL"

  # PREPARE test-data
  rm -rf test-data
  mkdir -p test-data
  cd test-data


  OLD_CONTAINER=$(docker run --link $POSTGRES -e "DB_URL=$DBURL" -e LOGIN=proxy -q -P --rm -d -v "$(pwd):/data" "ghcr.io/reddec/token-login:${PREV_VERSION}")
  OLD_ADMIN_PORT=$(docker inspect ${OLD_CONTAINER} -f '{{(index (index .NetworkSettings.Ports "8080/tcp") 0).HostPort}}')

  echo "Waiting for DB - 5s"
  sleep 5 # to let database initialize

  curl -H 'X-User: admin' -f -d "label=minimal" http://localhost:$OLD_ADMIN_PORT
  curl -H 'X-User: admin' -f -d "label=with+path&path=/**" http://localhost:$OLD_ADMIN_PORT
  curl -H 'X-User: admin' -f -d "label=with+headers" http://localhost:$OLD_ADMIN_PORT

  # add headers
  curl -H 'X-User: admin' -f -d "name=foo&value=bar&action=headers" "http://localhost:$OLD_ADMIN_PORT/token/3/"
  curl -H 'X-User: admin' -f -d "name=x&value=y&action=headers" "http://localhost:$OLD_ADMIN_PORT/token/3/"

  echo "Stopping old service"
  docker rm -f ${OLD_CONTAINER}

  if [ -f token-login.sqlite ]; then
    # just dump for history
    sqlite3 -table token-login.sqlite 'SELECT id, key_id, user, label, path, headers FROM token'
  fi

  echo "Starting new service"
  TARGET_CONTAINER=$(docker run --link $POSTGRES -e "DB_URL=$DBURL" -e LOGIN=proxy --pull never -q -P --rm -d -v "$(pwd):/data" "ghcr.io/reddec/token-login:${TARGET_VERSION}")
  TARGET_ADMIN_PORT=$(docker inspect ${TARGET_CONTAINER} -f '{{(index (index .NetworkSettings.Ports "8080/tcp") 0).HostPort}}')

  echo "Waiting for DB - 5s"
  sleep 5 # to let database initialize


  CONTENT="$(curl -H 'X-User: admin' -f "http://localhost:$TARGET_ADMIN_PORT/api/v1/tokens")"
  echo "$CONTENT" | jq -e '.[] | select(.id == 1 and .label == "minimal" and .path == "")'
  echo "$CONTENT" | jq -e '.[] | select(.id == 2 and .label == "with path" and .path == "/**")'
  echo "$CONTENT" | jq -e '.[] | select(.id == 3 and .label == "with headers" and .path == "")'

  echo "$CONTENT" | jq -e '.[] | select(.id == 3).headers | .[] | select(.name == "foo" and .value == "bar")'
  echo "$CONTENT" | jq -e '.[] | select(.id == 3).headers | .[] | select(.name == "x" and .value == "y")'

  docker rm -f ${TARGET_CONTAINER}
  cd ../
done

docker rm -f $POSTGRES