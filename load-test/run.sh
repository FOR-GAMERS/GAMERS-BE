#!/bin/bash
set -euo pipefail

# ─── GAMERS Load Test Runner ───
# Usage:
#   ./run.sh                        # 기본: team-invite 시나리오
#   ./run.sh team-invite            # Team 초대 시나리오
#   ./run.sh application            # 참가 신청 시나리오
#   ./run.sh contest-start          # 대회 시작 시나리오
#   ./run.sh tournament-create      # 토너먼트 생성 시나리오
#   ./run.sh full-flow              # 전체 플로우 시나리오
#   ./run.sh team-query             # 팀 조회 N+1 성능 시나리오
#   ./run.sh dashboard              # 대시보드만 실행 (Grafana + InfluxDB)
#   ./run.sh clean                  # 모든 컨테이너 & 볼륨 정리

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

SCENARIO="${1:-team-invite}"
COMPOSE="docker compose -f docker-compose.yaml"

# ─── 환경변수 기본값 ───
export LEADER_TOKEN="${LEADER_TOKEN:-}"
export LOGIN_EMAIL="${LOGIN_EMAIL:-}"
export LOGIN_PASSWORD="${LOGIN_PASSWORD:-}"
export INVITEE_TOKEN="${INVITEE_TOKEN:-your-invitee-jwt-token}"
export CONTEST_ID="${CONTEST_ID:-1}"
export BASE_URL="${BASE_URL:-http://gamers-app:8080/api}"

# ─── 네트워크 확인 ───
ensure_network() {
  if ! docker network inspect gamers-network >/dev/null 2>&1; then
    echo "[INFO] Creating gamers-network..."
    docker network create gamers-network
  fi
}

# ─── 대시보드 시작 ───
start_dashboard() {
  echo "============================================"
  echo "  Starting InfluxDB + Grafana Dashboard"
  echo "============================================"
  $COMPOSE up -d influxdb grafana
  echo ""
  echo "[OK] Grafana:  http://localhost:3001  (admin / admin)"
  echo "[OK] InfluxDB: http://localhost:8086"
  echo ""
}

# ─── k6 시나리오 실행 ───
run_scenario() {
  local script="$1"
  echo "============================================"
  echo "  Running scenario: ${script}"
  echo "============================================"
  echo "  BASE_URL:   ${BASE_URL}"
  echo "  CONTEST_ID: ${CONTEST_ID}"
  echo "============================================"
  echo ""

  $COMPOSE run --rm \
    -e K6_OUT="influxdb=http://influxdb:8086/k6" \
    -e LEADER_TOKEN="${LEADER_TOKEN}" \
    -e LOGIN_EMAIL="${LOGIN_EMAIL}" \
    -e LOGIN_PASSWORD="${LOGIN_PASSWORD}" \
    -e INVITEE_TOKEN="${INVITEE_TOKEN}" \
    -e CONTEST_ID="${CONTEST_ID}" \
    -e BASE_URL="${BASE_URL}" \
    -e USER_TOKENS="${USER_TOKENS:-}" \
    -e USER_IDS="${USER_IDS:-}" \
    -e CONTEST_IDS="${CONTEST_IDS:-}" \
    -e LEADER_TOKENS="${LEADER_TOKENS:-}" \
    -e INVITEE_IDS="${INVITEE_IDS:-}" \
    k6 run --out "influxdb=http://influxdb:8086/k6" "/scripts/scenario-${script}.js"
}

# ─── 정리 ───
cleanup() {
  echo "[INFO] Stopping all containers and removing volumes..."
  $COMPOSE down -v
  echo "[OK] Cleaned up."
}

# ─── Main ───
ensure_network

case "$SCENARIO" in
  dashboard)
    start_dashboard
    echo "Dashboard is running. Press Ctrl+C to stop."
    echo "Run tests in another terminal: ./run.sh team-invite"
    $COMPOSE logs -f grafana
    ;;
  clean)
    cleanup
    ;;
  team-invite|application|contest-start|tournament-create|full-flow|team-query)
    start_dashboard
    sleep 3  # InfluxDB가 준비될 때까지 대기
    run_scenario "$SCENARIO"
    echo ""
    echo "============================================"
    echo "  Test Complete!"
    echo "  View results: http://localhost:3001"
    echo "============================================"
    ;;
  *)
    echo "Unknown scenario: $SCENARIO"
    echo ""
    echo "Available scenarios:"
    echo "  team-invite         Team 초대 부하 테스트"
    echo "  application         참가 신청 부하 테스트"
    echo "  contest-start       대회 시작 (배치 INSERT) 테스트"
    echo "  tournament-create   토너먼트 생성 테스트"
    echo "  full-flow           전체 플로우 E2E 테스트"
    echo "  team-query          팀 조회 N+1 성능 테스트"
    echo "  dashboard           대시보드만 실행"
    echo "  clean               정리"
    exit 1
    ;;
esac
