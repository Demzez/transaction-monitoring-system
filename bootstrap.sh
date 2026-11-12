#!/bin/sh
set -eu

# Execute from the root of transaction-monitoring-system.
#cd "$(dirname "$0")/.."

command -v docker >/dev/null || { echo "Docker CLI not found" >&2; exit 1; }
docker compose version >/dev/null

for net in tms-ci tms-runtime; do
    if ! docker network inspect "$net" >/dev/null 2>&1; then
        docker network create "$net"
    fi
done

POSTGRES_CONTAINER=postgres_2d2
if ! docker inspect "$POSTGRES_CONTAINER" >/dev/null 2>&1; then
    echo "Existing container $POSTGRES_CONTAINER not found." >&2
    exit 1
fi

if ! docker network inspect tms-runtime --format '{{range $id, $c := .Containers}}{{println $c.Name}}{{end}}' | grep -Fxq "$POSTGRES_CONTAINER"; then
    docker network connect --alias postgres tms-runtime "$POSTGRES_CONTAINER"
fi

# Check the chosen alias is present; do not change live network unnecessarily.
if ! docker inspect -f '{{range .NetworkSettings.Networks}}{{range .Aliases}}{{println .}}{{end}}{{end}}' "$POSTGRES_CONTAINER" | grep -Fxq postgres; then
    echo "Network alias postgres is missing on $POSTGRES_CONTAINER." >&2
    echo "Reconnect it once using: docker network disconnect tms-runtime $POSTGRES_CONTAINER" >&2
    echo "then: docker network connect --alias postgres tms-runtime $POSTGRES_CONTAINER" >&2
    exit 1
fi

SOCKET=/var/run/docker.sock
if [ ! -S "$SOCKET" ]; then
    echo "Docker Unix socket not found: $SOCKET" >&2
    exit 1
fi
DOCKER_GID="$(stat -c '%g' "$SOCKET")"
export DOCKER_GID

docker compose -f jenkins/compose.yaml up -d --build

echo "Jenkins URL: http://localhost:8081"
echo "Verify Docker access: docker exec local-jenkins docker ps"
echo "Initial admin password: docker exec local-jenkins cat /var/jenkins_home/secrets/initialAdminPassword"
