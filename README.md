**SISTEMAS DISTRIBUIDOS — TAREA 3 — 2026-1**
<br>

**INTEGRANTES:**
  - Kris Casanga — 202021069-1
  - Gonzalo Severín — 202073088-1
  - Luz Vilches — 202273033-1
<br>

**INSTRUCCIONES DE USO:**

Requisito:
  - Tener instalado Go 1.2 o superior.
  - Poseer una cuenta del Departamento de Informática.

<br>
Se uso Rest

ya main, server, cliente y estado, deberian estar listos los probe localmente, intrucciones solo tiene codigo para que compile no implementa nada real. Si agregan funciones adicionales o otras cosas verifiquen si deben agregar en alguno de los archivos que trabaje, como no hay logica detras no hacen nada mas que mostrar que esta conectado, no he subido nada a las mv porque no tiene sentido todavia:

lo que probe:
```cmd
cd expendedora

go mod init expendedora

go build -o ../bin/expendedora .

respuesta:
nada

cd ..
./bin/expendedora 1 1 8101 "localhost:8102,localhost:8103" &
./bin/expendedora 1 2 8102 "localhost:8101,localhost:8103" &
./bin/expendedora 1 3 8103 "localhost:8101,localhost:8102" &

respuesta:
[1] 7429
[2] 7430
[3] 7431

2026/06/10 22:57:26 [M1P2] Escuchando en puerto 8102
2026/06/10 22:57:26 [M1P3] Escuchando en puerto 8103
2026/06/10 22:57:26 [M1P1] Escuchando en puerto 8101
2026/06/10 22:57:27 [M1P1] Esperando peers...
2026/06/10 22:57:27 [M1P3] Esperando peers...
2026/06/10 22:57:27 Esperando peer localhost:8102...
2026/06/10 22:57:27 Esperando peer localhost:8101...
2026/06/10 22:57:27 [M1P2] Esperando peers...
2026/06/10 22:57:27 Esperando peer localhost:8101...
2026/06/10 22:57:27 Peer localhost:8101 listo
2026/06/10 22:57:27 Peer localhost:8102 listo
2026/06/10 22:57:27 Esperando peer localhost:8103...
2026/06/10 22:57:27 Esperando peer localhost:8102...
2026/06/10 22:57:27 Peer localhost:8101 listo
2026/06/10 22:57:27 Esperando peer localhost:8103...
2026/06/10 22:57:27 Peer localhost:8103 listo
2026/06/10 22:57:27 [M1P2] Finalizada espera de peers
2026/06/10 22:57:27 Peer localhost:8102 listo
2026/06/10 22:57:27 [M1P3] Finalizada espera de peers
2026/06/10 22:57:27 Peer localhost:8103 listo
2026/06/10 22:57:27 [M1P1] Finalizada espera de peers
2026/06/10 22:57:27 [M1P3] Inventario cargado
2026/06/10 22:57:27 [M1P1] Inventario cargado
2026/06/10 22:57:27 [M1P2] Inventario cargado
2026/06/10 22:57:27 [M1P1] Ejecutando instrucciones desde instrucciones/proceso_1.txt
2026/06/10 22:57:27 [M1P1] Instrucciones terminadas. Proceso sigue escuchando...
2026/06/10 22:57:27 [M1P2] No se encontró archivo para proceso 2
2026/06/10 22:57:27 [M1P3] No se encontró archivo para proceso 3

(LO DE NO SE ENCONTRO ES PORQUE SOLO ESTA EL EJEMPLO QUE SALE EN EL ENUNCIADO Y HABRIA QUE CREAR 2 ARCHIVOS MAS PARA PROBAR)

en otra terminal:
curl localhost:8101/ping
curl localhost:8101/estado

respuesta:
pong
{"inventario":[{"nombre":"manzana","cantidad":100},{"nombre":"naranja","cantidad":10}],"malicioso":false,"vetos":{}}
```
