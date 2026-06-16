# INF-343 Sistemas Distribuidos - Tarea 3

## Integrantes

- Kris Casanga - 202021069-1
- Gonzalo Severin - 202073088-1
- Luz Vilches - 202273033-1

## Arquitectura

La solucion usa procesos escritos en Go que se comunican mediante una API REST sobre HTTP. Cada expendedora corresponde a un proceso con puerto propio. Los procesos mantienen estado local con mutex y replican inventario y lista de vetos a los demas procesos mediante endpoints REST.

Cada maquina ejecuta la misma cantidad de procesos. Los puertos se calculan con:

```text
PORT_BASE + numero_maquina * 100 + numero_proceso
```

Por defecto `PORT_BASE=8100`, por lo que `M1P1` escucha en `8201`, `M2P1` en `8301` y `M3P1` en `8401`.

## Requisitos

- Ubuntu en las maquinas virtuales.
- Go 1.22 o superior.
- `curl`.
- Acceso de red entre las tres maquinas por los puertos usados.

## Guia rapida para probar en las VMs

Para una prueba distribuida real se deben usar las tres maquinas al mismo
tiempo. Cada maquina ejecuta el mismo repositorio, pero se inicia con un numero
distinto:

- M1: `10.10.28.17`
- M2: `10.10.28.18`
- M3: `10.10.28.19`

En cada VM, entrar a la carpeta del repositorio y preparar las variables:

```bash
cd Tarea3-INF343-Grupo3
export MAQUINA1_HOST=10.10.28.17
export MAQUINA2_HOST=10.10.28.18
export MAQUINA3_HOST=10.10.28.19
sed -i 's/\r$//' script.sh
chmod +x script.sh
```

Luego iniciar una terminal SSH para cada VM y ejecutar:

```bash
# En 10.10.28.17
./script.sh 1 3

# En 10.10.28.18
./script.sh 2 3

# En 10.10.28.19
./script.sh 3 3
```

El segundo parametro es la cantidad de procesos por maquina. Como existen
`proceso_1.txt`, `proceso_2.txt` y `proceso_3.txt`, la prueba completa usa `3`.
Para una prueba minima se puede usar `1`, pero solo se ejecutara
`proceso_1.txt`.

Para revisar que quedo corriendo:

```bash
./script.sh 1 ESTADO 1
curl http://localhost:8201/estado
ls -lh logs/
```

Para detener los procesos locales de una VM:

```bash
./script.sh <NUMERO_DE_MAQUINA> KILLALL
```

## Configuracion en las tres maquinas

Antes de ejecutar, exportar las IP o nombres DNS de las tres VMs en cada maquina:

```bash
export MAQUINA1_HOST=10.10.28.17
export MAQUINA2_HOST=10.10.28.18
export MAQUINA3_HOST=10.10.28.19
```

Si se quiere cambiar la base de puertos:

```bash
export PORT_BASE=8100
```

Dar permiso de ejecucion al script:

```bash
sed -i 's/\r$//' script.sh
chmod +x script.sh
```

## Inicializacion

En cada VM ejecutar el mismo numero de procesos, cambiando solo el numero de maquina:

```bash
./script.sh 1 <CANTIDAD_DE_PROCESOS>
./script.sh 2 <CANTIDAD_DE_PROCESOS>
./script.sh 3 <CANTIDAD_DE_PROCESOS>
```

Ejemplo con un proceso por maquina:

```bash
./script.sh 1 1
./script.sh 2 1
./script.sh 3 1
```

Cada proceso espera hasta 2 segundos por los peers y luego queda disponible para procesar instrucciones y recibir actualizaciones.

## Operaciones

Restaurar un proceso:

```bash
./script.sh <NUMERO_DE_MAQUINA> RESTAURAR <NUMERO_DE_ID_DEL_TXT>
```

Matar un proceso:

```bash
./script.sh <NUMERO_DE_MAQUINA> MATAR <NUMERO_DE_ID_DEL_TXT>
```

Matar todos los procesos de la maquina:

```bash
./script.sh <NUMERO_DE_MAQUINA> KILLALL
```

Infectar o desinfectar los procesos locales de la maquina:

```bash
./script.sh INFECTAR
```

Ver estado de un proceso:

```bash
./script.sh <NUMERO_DE_MAQUINA> ESTADO <NUMERO_DE_ID_DEL_TXT>
```

## Formato de archivos

Los inventarios deben estar en `inventario/*.json`:

```json
[
  {"nombre": "manzana", "cantidad": 100},
  {"nombre": "naranja", "cantidad": 10}
]
```

Las instrucciones deben estar en `instrucciones/*_<ID>.txt`, por ejemplo `proceso_1.txt`:

```text
VETAR jack
COMPRAR jack manzana 10
COMPRAR anna dewitt manzana 15
PERDONAR jack
```

## Logs

Los logs se escriben en `logs/` con los formatos pedidos:

- `inventario_M<maquina>P<proceso>.log`
- `vetos_M<maquina>P<proceso>.log`

## Recuperacion

Al restaurar, el proceso consulta `/snapshot` a sus peers durante hasta 3 segundos. Luego elige el inventario con mayor cantidad de replicas iguales. Si no existe una mayoria estricta mayor a dos tercios de las respuestas recibidas, el proceso termina con error y no entra al sistema.

Si un proceso esta infectado, responde snapshots con inventario corrupto. Esto permite probar que la restauracion rechaza estados sin consenso suficiente.

## Consideraciones

- La solucion replica listas completas de inventario y vetos.
- Las operaciones locales sobre inventario y vetos usan mutex.
- Existe una sincronizacion periodica cada 5 segundos hacia los peers.
- `INFECTAR` alterna el modo malicioso de los procesos locales. Si se ejecuta otra vez, vuelven al modo normal.

## Uso de asistencia por IA

Se utilizo asistencia por IA para apoyar la implementacion de la logica REST, el script de ejecucion, la recuperacion por mayoria y la redaccion de este README. El codigo fue revisado para eliminar comentarios automaticos o texto no relacionado con la entrega.
