#!/bin/sh
set -eu

# Spring listens internally; nginx is the public :8080 for the Worker probe/proxy.
export SERVER_PORT=8081

java -jar /app/app.jar &
JAVA_PID=$!

ready=0
i=0
while [ "$i" -lt 90 ]; do
	if curl -sf "http://127.0.0.1:8081/api/v1/health" >/dev/null; then
		ready=1
		break
	fi
	# Exit early if JVM died during boot.
	if ! kill -0 "$JAVA_PID" 2>/dev/null; then
		echo "Spring Boot exited before becoming healthy" >&2
		wait "$JAVA_PID" || true
		exit 1
	fi
	i=$((i + 1))
	sleep 2
done

if [ "$ready" -ne 1 ]; then
	echo "Spring Boot did not become ready on :8081" >&2
	kill "$JAVA_PID" 2>/dev/null || true
	exit 1
fi

exec nginx -g "daemon off;"
