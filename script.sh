#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="$ROOT_DIR/bin"
RUN_DIR="$ROOT_DIR/.run"
LOG_DIR="$ROOT_DIR/logs"
APP_DIR="$ROOT_DIR/expendedora"
PORT_BASE="${PORT_BASE:-8100}"

mkdir -p "$BIN_DIR" "$RUN_DIR" "$LOG_DIR"

usage() {
  cat <<USAGE
Uso:
  ./script.sh <NUMERO_DE_MAQUINA> <CANTIDAD_DE_PROCESOS>
  ./script.sh <NUMERO_DE_MAQUINA> RESTAURAR <NUMERO_DE_ID_DEL_TXT>
  ./script.sh <NUMERO_DE_MAQUINA> MATAR <NUMERO_DE_ID_DEL_TXT>
  ./script.sh <NUMERO_DE_MAQUINA> KILLALL
  ./script.sh <NUMERO_DE_MAQUINA> ESTADO <NUMERO_DE_ID_DEL_TXT>
  ./script.sh INFECTAR

Variables opcionales:
  MAQUINA1_HOST, MAQUINA2_HOST, MAQUINA3_HOST para indicar IP/host de cada VM.
  PORT_BASE para cambiar la base de puertos. Por defecto usa 8100.
USAGE
}

host_maquina() {
  local maquina="$1"
  local var="MAQUINA${maquina}_HOST"
  echo "${!var:-localhost}"
}

puerto_proceso() {
  local maquina="$1"
  local proceso="$2"
  echo $((PORT_BASE + maquina * 100 + proceso))
}

cantidad_procesos() {
  if [[ -n "${NUM_PROCESOS:-}" ]]; then
    echo "$NUM_PROCESOS"
  elif [[ -f "$RUN_DIR/cantidad_procesos" ]]; then
    cat "$RUN_DIR/cantidad_procesos"
  else
    echo "1"
  fi
}

peers_para() {
  local maquina_actual="$1"
  local proceso_actual="$2"
  local cantidad="$3"
  local peers=()

  for maquina in 1 2 3; do
    local host
    host="$(host_maquina "$maquina")"
    for proceso in $(seq 1 "$cantidad"); do
      if [[ "$maquina" == "$maquina_actual" && "$proceso" == "$proceso_actual" ]]; then
        continue
      fi
      peers+=("${host}:$(puerto_proceso "$maquina" "$proceso")")
    done
  done

  local IFS=,
  echo "${peers[*]}"
}

build_app() {
  echo "Compilando expendedora..."
  (cd "$APP_DIR" && go build -o "$BIN_DIR/expendedora" .)
  echo "Compilacion lista: $BIN_DIR/expendedora"
}

iniciar_proceso() {
  local maquina="$1"
  local proceso="$2"
  local cantidad="$3"
  local modo="${4:-}"
  local puerto peers pid_file

  puerto="$(puerto_proceso "$maquina" "$proceso")"
  peers="$(peers_para "$maquina" "$proceso" "$cantidad")"
  pid_file="$RUN_DIR/M${maquina}P${proceso}.pid"
  if [[ -f "$pid_file" ]] && kill -0 "$(cat "$pid_file")" 2>/dev/null; then
    echo "M${maquina}P${proceso} ya está corriendo con PID $(cat "$pid_file")"
    return
  fi

  if [[ "$modo" == "RESTAURAR" ]]; then
    echo "Iniciando M${maquina}P${proceso} en modo RESTAURAR"
    echo "  puerto=${puerto}"
    echo "  peers=${peers}"
    (cd "$APP_DIR" && "$BIN_DIR/expendedora" "$maquina" "$proceso" "$puerto" "$peers" RESTAURAR & echo $! > "$pid_file")
  else
    echo "Iniciando M${maquina}P${proceso}"
    echo "  puerto=${puerto}"
    echo "  peers=${peers}"
    (cd "$APP_DIR" && "$BIN_DIR/expendedora" "$maquina" "$proceso" "$puerto" "$peers" & echo $! > "$pid_file")
  fi

  echo "M${maquina}P${proceso} iniciado en puerto ${puerto} con PID $(cat "$pid_file")"
}

matar_proceso() {
  local maquina="$1"
  local proceso="$2"
  local pid_file="$RUN_DIR/M${maquina}P${proceso}.pid"

  if [[ ! -f "$pid_file" ]]; then
    echo "No hay PID registrado para M${maquina}P${proceso}"
    return
  fi

  local pid
  pid="$(cat "$pid_file")"
  if kill -0 "$pid" 2>/dev/null; then
    kill "$pid"
    echo "M${maquina}P${proceso} detenido"
  else
    echo "M${maquina}P${proceso} no estaba corriendo"
  fi
  rm -f "$pid_file"
}

estado_proceso() {
  local maquina="$1"
  local proceso="$2"
  local host puerto
  host="$(host_maquina "$maquina")"
  puerto="$(puerto_proceso "$maquina" "$proceso")"
  curl -s "http://${host}:${puerto}/estado"
  echo
}

infectar_locales() {
  local cantidad maquina
  cantidad="$(cantidad_procesos)"
  if [[ -f "$RUN_DIR/maquina_local" ]]; then
    maquina="$(cat "$RUN_DIR/maquina_local")"
  else
    maquina="${MAQUINA_LOCAL:-1}"
  fi

  for proceso in $(seq 1 "$cantidad"); do
    local puerto
    puerto="$(puerto_proceso "$maquina" "$proceso")"
    curl -s -X POST "http://localhost:${puerto}/infectar" >/dev/null || true
  done
  echo "Procesos locales de M${maquina} alternaron modo malicioso"
}

if [[ "$#" -lt 1 ]]; then
  usage
  exit 1
fi

if [[ "$1" == "INFECTAR" ]]; then
  infectar_locales
  exit 0
fi

if [[ "$#" -lt 2 ]]; then
  usage
  exit 1
fi

MAQUINA="$1"
ACCION="$2"

case "$ACCION" in
  ''|*[!0-9]*)
    case "$ACCION" in
      RESTAURAR)
        [[ "$#" -eq 3 ]] || { usage; exit 1; }
        build_app
        iniciar_proceso "$MAQUINA" "$3" "$(cantidad_procesos)" "RESTAURAR"
        ;;
      MATAR)
        [[ "$#" -eq 3 ]] || { usage; exit 1; }
        matar_proceso "$MAQUINA" "$3"
        ;;
      KILLALL)
        for pid_file in "$RUN_DIR"/M"${MAQUINA}"P*.pid; do
          [[ -e "$pid_file" ]] || continue
          proceso="${pid_file##*P}"
          proceso="${proceso%.pid}"
          matar_proceso "$MAQUINA" "$proceso"
        done
        ;;
      ESTADO)
        [[ "$#" -eq 3 ]] || { usage; exit 1; }
        estado_proceso "$MAQUINA" "$3"
        ;;
      *)
        usage
        exit 1
        ;;
    esac
    ;;
  *)
    CANTIDAD="$ACCION"
    echo "$CANTIDAD" > "$RUN_DIR/cantidad_procesos"
    echo "$MAQUINA" > "$RUN_DIR/maquina_local"
    build_app
    for proceso in $(seq 1 "$CANTIDAD"); do
      iniciar_proceso "$MAQUINA" "$proceso" "$CANTIDAD"
    done
    ;;
esac
