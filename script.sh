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
  ./script.sh <NUMERO_DE_MAQUINA> LOGS <NUMERO_DE_ID_DEL_TXT>
  ./script.sh <NUMERO_DE_MAQUINA> RUNTIME <NUMERO_DE_ID_DEL_TXT> [LINEAS]
  ./script.sh INFECTAR

Variables opcionales:
  MAQUINA1_HOST, MAQUINA2_HOST, MAQUINA3_HOST para indicar IP/host de cada VM.
  PORT_BASE para cambiar la base de puertos. Por defecto usa 8100.
USAGE
}

print_header() {
  local titulo="$1"
  echo
  echo "============================================================"
  echo " $titulo"
  echo "============================================================"
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
  print_header "Compilando expendedora"
  (cd "$APP_DIR" && go build -o "$BIN_DIR/expendedora" .)
  echo "OK  Binario listo: $BIN_DIR/expendedora"
}

iniciar_proceso() {
  local maquina="$1"
  local proceso="$2"
  local cantidad="$3"
  local modo="${4:-}"
  local puerto peers pid_file runtime_log

  puerto="$(puerto_proceso "$maquina" "$proceso")"
  peers="$(peers_para "$maquina" "$proceso" "$cantidad")"
  pid_file="$RUN_DIR/M${maquina}P${proceso}.pid"
  runtime_log="$LOG_DIR/runtime_M${maquina}P${proceso}.log"

  if [[ -f "$pid_file" ]] && kill -0 "$(cat "$pid_file")" 2>/dev/null; then
    echo "OK  M${maquina}P${proceso} ya esta corriendo con PID $(cat "$pid_file")"
    return
  fi

  if [[ "$modo" == "RESTAURAR" ]]; then
    print_header "Iniciando M${maquina}P${proceso} en modo RESTAURAR"
    echo "  Puerto : ${puerto}"
    echo "  Peers  : ${peers}"
    echo "  Runtime: ${runtime_log}"
    (cd "$ROOT_DIR" && "$BIN_DIR/expendedora" "$maquina" "$proceso" "$puerto" "$peers" RESTAURAR >> "$runtime_log" 2>&1 & echo $! > "$pid_file")
  else
    print_header "Iniciando M${maquina}P${proceso}"
    echo "  Puerto : ${puerto}"
    echo "  Peers  : ${peers}"
    echo "  Runtime: ${runtime_log}"
    (cd "$ROOT_DIR" && "$BIN_DIR/expendedora" "$maquina" "$proceso" "$puerto" "$peers" >> "$runtime_log" 2>&1 & echo $! > "$pid_file")
  fi

  echo "OK  M${maquina}P${proceso} iniciado con PID $(cat "$pid_file")"
}

matar_proceso() {
  local maquina="$1"
  local proceso="$2"
  local pid_file="$RUN_DIR/M${maquina}P${proceso}.pid"

  if [[ ! -f "$pid_file" ]]; then
    echo "INFO No hay PID registrado para M${maquina}P${proceso}"
    return
  fi

  local pid
  pid="$(cat "$pid_file")"
  if kill -0 "$pid" 2>/dev/null; then
    kill "$pid"
    echo "OK  M${maquina}P${proceso} detenido"
  else
    echo "INFO M${maquina}P${proceso} no estaba corriendo"
  fi
  rm -f "$pid_file"
}

estado_proceso() {
  local maquina="$1"
  local proceso="$2"
  local host puerto respuesta
  host="$(host_maquina "$maquina")"
  puerto="$(puerto_proceso "$maquina" "$proceso")"
  respuesta="$(curl -s "http://${host}:${puerto}/estado")"

  print_header "Estado M${maquina}P${proceso}"
  if command -v python3 >/dev/null 2>&1; then
    printf '%s\n' "$respuesta" | python3 -m json.tool
  else
    printf '%s\n' "$respuesta"
  fi
}

mostrar_logs() {
  local maquina="$1"
  local proceso="$2"
  local inventario_log="$LOG_DIR/inventario_M${maquina}P${proceso}.log"
  local vetos_log="$LOG_DIR/vetos_M${maquina}P${proceso}.log"

  print_header "Log de inventario M${maquina}P${proceso}"
  if [[ -s "$inventario_log" ]]; then
    awk -F ' \\| ' '
      BEGIN {
        printf "  %-4s %-46s %-12s\n", "#", "Instruccion", "Resultado"
        printf "  %-4s %-46s %-12s\n", "----", "----------------------------------------------", "------------"
      }
      {
        resultado = $2
        if (resultado == "") {
          resultado = "-"
        }
        printf "  %-4d %-46s %-12s\n", NR, $1, resultado
      }
    ' "$inventario_log"
  else
    echo "  Sin registros de inventario."
  fi

  print_header "Lista de vetos M${maquina}P${proceso}"
  if [[ -s "$vetos_log" ]]; then
    awk '
      BEGIN {
        printf "  %-34s %-8s\n", "Persona", "Counter"
        printf "  %-34s %-8s\n", "----------------------------------", "--------"
      }
      /^VETADO / {
        counter = $NF
        persona = $0
        sub(/^VETADO /, "", persona)
        sub(" " counter "$", "", persona)
        printf "  %-34s %-8s\n", persona, counter
      }
    ' "$vetos_log"
  else
    echo "  Sin personas vetadas."
  fi
}

mostrar_runtime() {
  local maquina="$1"
  local proceso="$2"
  local lineas="${3:-40}"
  local runtime_log="$LOG_DIR/runtime_M${maquina}P${proceso}.log"

  print_header "Comunicacion interna M${maquina}P${proceso} (ultimas ${lineas} lineas)"
  if [[ -s "$runtime_log" ]]; then
    tail -n "$lineas" "$runtime_log"
  else
    echo "  Sin registros runtime."
    echo "  Si el proceso fue iniciado antes de este cambio, reinicialo."
  fi
}

infectar_locales() {
  local cantidad maquina
  cantidad="$(cantidad_procesos)"
  if [[ -f "$RUN_DIR/maquina_local" ]]; then
    maquina="$(cat "$RUN_DIR/maquina_local")"
  else
    maquina="${MAQUINA_LOCAL:-1}"
  fi

  print_header "Alternando modo malicioso en M${maquina}"
  for proceso in $(seq 1 "$cantidad"); do
    local puerto
    puerto="$(puerto_proceso "$maquina" "$proceso")"
    curl -s -X POST "http://localhost:${puerto}/infectar" >/dev/null || true
    echo "OK  M${maquina}P${proceso} alternado"
  done
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
      LOGS)
        [[ "$#" -eq 3 ]] || { usage; exit 1; }
        mostrar_logs "$MAQUINA" "$3"
        ;;
      RUNTIME)
        [[ "$#" -eq 3 || "$#" -eq 4 ]] || { usage; exit 1; }
        mostrar_runtime "$MAQUINA" "$3" "${4:-40}"
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
