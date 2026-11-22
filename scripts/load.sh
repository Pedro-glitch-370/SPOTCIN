#!/usr/bin/env bash
# Criei esse script pra mostrar os clientes funcionando em paralelo
# Dispara N clientes de terminal em paralelo contra o back

set -euo pipefail

BASE="${1:-http://localhost:8080}"
BACKEND_DIR="$(cd "$(dirname "$0")/../backend" && pwd)"

echo "Catálogo em $BASE"
SONGS_JSON=$(curl -sS "$BASE/api/songs")
COUNT=$(echo "$SONGS_JSON" | python -c "import sys,json; print(len(json.load(sys.stdin)))")
echo "$COUNT músicas no catálogo"

if [ "$COUNT" -lt 1 ]; then
  echo "Catálogo vazio, rode o backend com devseed/audio populado" >&2
  exit 1
fi

# extrai os IDs das duas músicas
IDS=$(echo "$SONGS_JSON" | python -c "
import sys, json
songs = json.load(sys.stdin)
for s in songs[:2]:
    print(s['id'])
")
ID_A=$(echo "$IDS" | sed -n '1p')
ID_B=$(echo "$IDS" | sed -n '2p')
[ -z "$ID_B" ] && ID_B="$ID_A"

echo "Música A: $ID_A"
echo "Música B: $ID_B"

cd "$BACKEND_DIR"

echo "Disparando 4 clientes em paralelo (logs em /tmp/spotcin-client-*.log)"

go run ./cmd/client -base "$BASE" -song "$ID_A" -user "load-a1" -drop-after 2097152 > /tmp/spotcin-client-a1.log 2>&1 &
PID_A1=$!

go run ./cmd/client -base "$BASE" -song "$ID_A" -user "load-a2" -jitter 100ms > /tmp/spotcin-client-a2.log 2>&1 &
PID_A2=$!

go run ./cmd/client -base "$BASE" -song "$ID_B" -user "load-b1" -drop-after 1048576 > /tmp/spotcin-client-b1.log 2>&1 &
PID_B1=$!

go run ./cmd/client -base "$BASE" -song "$ID_B" -user "load-b2" -start 5 -drop-after 1048576 > /tmp/spotcin-client-b2.log 2>&1 &
PID_B2=$!

echo "PIDs: a1=$PID_A1 a2=$PID_A2 b1=$PID_B1 b2=$PID_B2"
echo "Aguardando término (Ctrl+C cancela todos). Toma cerca de 283 segundos."

# espera todos
# se um falhar, mata os outros também
trap 'kill $PID_A1 $PID_A2 $PID_B1 $PID_B2 2>/dev/null || true' INT TERM
wait $PID_A1 $PID_A2 $PID_B1 $PID_B2 || true

echo "Fim. Resumo dos clientes:"
for f in /tmp/spotcin-client-*.log; do
  echo "--- $f"
  grep -E "posição final|fim:|erro|ERRO" "$f" | tail -3 || true
done

echo
echo "No terminal do servidor, deve haver 4 request_id diferentes"
echo "com status:206, ranges diferentes e com durações sobrepostas."