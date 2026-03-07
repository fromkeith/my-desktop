#!/bin/bash
set -e

echo "custom start"

install -D -m 400 /tmp/pwfile /mongot-community/pwfile

exec /mongot-community/mongot --config /mongot-community/config.default.yml
