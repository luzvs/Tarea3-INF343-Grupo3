package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

// main inicializa el servidor, carga o recupera estado y ejecuta instrucciones.
func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	if len(os.Args) < 5 {
		fmt.Println(
			"Uso: ./expendedora <num_maquina> <id_proceso> <puerto> <peers>",
		)
		os.Exit(1)
	}

	numMaquina, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("num_maquina invalido: %v", err)
	}

	idProceso, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatalf("id_proceso invalido: %v", err)
	}

	puerto := os.Args[3]
	peers := ParsearPeers(os.Args[4])
	modoRestaurar := len(os.Args) >= 6 && os.Args[5] == "RESTAURAR"

	log.Printf(
		"[M%dP%d] Configuracion: puerto=%s peers=%v restaurar=%v",
		numMaquina,
		idProceso,
		puerto,
		peers,
		modoRestaurar,
	)

	estado := NuevoEstado(
		numMaquina,
		idProceso,
	)

	srv := NuevoServidor(
		puerto,
		estado,
	)

	go func() {
		if err := srv.Iniciar(); err != nil {
			log.Fatalf(
				"[M%dP%d] Error iniciando servidor: %v",
				numMaquina,
				idProceso,
				err,
			)
		}
	}()


	time.Sleep(300 * time.Millisecond)

	log.Printf(
		"[M%dP%d] Esperando peers...",
		numMaquina,
		idProceso,
	)

	EsperarPeers(
		peers,
		2,
	)

	log.Printf(
		"[M%dP%d] Finalizada espera de peers",
		numMaquina,
		idProceso,
	)


	if modoRestaurar {
		log.Printf(
			"[M%dP%d] Restaurando estado desde peers...",
			numMaquina,
			idProceso,
		)

		if err := RecuperarEstado(estado, peers); err != nil {
			log.Fatalf(
				"[M%dP%d] Error recuperando estado: %v",
				numMaquina,
				idProceso,
				err,
			)
		}

		log.Printf(
			"[M%dP%d] Estado recuperado",
			numMaquina,
			idProceso,
		)
	} else {
		if err := estado.CargarInventarioAleatorio(
			"inventario",
		); err != nil {

			log.Fatalf(
				"[M%dP%d] Error cargando inventario: %v",
				numMaquina,
				idProceso,
				err,
			)
		}

		log.Printf(
			"[M%dP%d] Inventario cargado",
			numMaquina,
			idProceso,
		)
	}

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			BroadcastSnapshot(peers, estado)
		}
	}()

	archivoInstr := BuscarArchivo(
		"instrucciones",
		idProceso,
	)

	if archivoInstr == "" {
		log.Printf(
			"[M%dP%d] No se encontró archivo para proceso %d",
			numMaquina,
			idProceso,
			idProceso,
		)
		select {}
	}

	log.Printf(
		"[M%dP%d] Ejecutando instrucciones desde %s",
		numMaquina,
		idProceso,
		archivoInstr,
	)


	ejecutor := NuevoEjecutor(
		estado,
		peers,
		numMaquina,
		idProceso,
	)

	if err := ejecutor.EjecutarArchivo(
		archivoInstr,
	); err != nil {

		log.Fatalf(
			"[M%dP%d] Error ejecutando instrucciones: %v",
			numMaquina,
			idProceso,
			err,
		)
	}

	log.Printf(
		"[M%dP%d] Instrucciones terminadas. Proceso sigue escuchando...",
		numMaquina,
		idProceso,
	)

	select {}
}
