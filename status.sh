#!/bin/sh
set -eu

echo '=== Docker services ==='
docker ps --format 'table {{.Names}}	{{.Image}}	{{.Ports}}'
echo '=== Application health ==='
docker inspect -f 'Health={{.State.Health.Status}} Image={{.Image}}' tms-app
echo '=== HTTP listener ==='
docker exec tms-app nc -z 127.0.0.1 8082
echo '=== TCP listener ==='
docker exec tms-app nc -z 127.0.0.1 8084
echo '=== Jenkins Docker socket access ==='
docker exec local-jenkins docker ps --format '{{.Names}}'
