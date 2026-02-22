#!/bin/sh
if [ -n "$LITESTREAM_REPLICA_URL" ]; then
  # Restore DB from replica if it exists (first deploy / data loss recovery)
  litestream restore -if-replica-exists -config /app/litestream.yml /app/pb_data/data.db
  # Start PocketBase under Litestream supervision
  exec litestream replicate -exec "/app/ender serve --http=0.0.0.0:8090" -config /app/litestream.yml
else
  exec /app/ender serve --http=0.0.0.0:8090
fi
