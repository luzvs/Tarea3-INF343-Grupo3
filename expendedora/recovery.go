package main

import (
	"fmt"
	"log"
	"time"
)

// RecuperarEstado solicita snapshots y recupera usando mayoria de inventarios.
func RecuperarEstado(estado *Estado, peers []string) error {
	deadline := time.After(3 * time.Second)
	respuestas := make(chan SnapshotEstado, len(peers))
	log.Printf("[M%dP%d] Recuperacion iniciada; consultando %d peers", estado.NumMaquina, estado.IdProceso, len(peers))

	for _, peer := range peers {
		go func(p string) {
			snapshot, err := ObtenerSnapshot(p)
			if err != nil {
				log.Printf("No se pudo recuperar desde %s: %v", p, err)
				return
			}
			respuestas <- snapshot
		}(peer)
	}

	var snapshots []SnapshotEstado
	esperados := len(peers)

	for len(snapshots) < esperados {
		select {
		case snapshot := <-respuestas:
			snapshots = append(snapshots, snapshot)
			log.Printf("[M%dP%d] Recuperacion: respuesta %d/%d recibida", estado.NumMaquina, estado.IdProceso, len(snapshots), esperados)
		case <-deadline:
			log.Printf("[M%dP%d] Recuperacion: timeout de 3 segundos con %d respuestas", estado.NumMaquina, estado.IdProceso, len(snapshots))
			return aplicarMayoritaria(estado, snapshots)
		}
	}

	return aplicarMayoritaria(estado, snapshots)
}

// aplicarMayoritaria valida que mas de dos tercios compartan el mismo inventario.
func aplicarMayoritaria(estado *Estado, snapshots []SnapshotEstado) error {
	if len(snapshots) == 0 {
		return fmt.Errorf("no se recibieron estados para recuperar")
	}

	conteo := make(map[string]int)
	representante := make(map[string]SnapshotEstado)

	for _, snapshot := range snapshots {
		if EsInventarioCorrupto(snapshot.Inventario) {
			log.Printf("Recuperacion: snapshot corrupto recibido; cuenta contra el consenso")
			continue
		}

		clave := ClaveInventario(snapshot.Inventario)
		conteo[clave]++
		representante[clave] = snapshot
		log.Printf("Recuperacion: inventario candidato=%s conteo=%d", clave, conteo[clave])
	}

	if len(conteo) == 0 {
		return fmt.Errorf("no se recibieron inventarios validos para recuperar")
	}

	var mejorClave string
	var mejorCantidad int
	for clave, cantidad := range conteo {
		if cantidad > mejorCantidad {
			mejorClave = clave
			mejorCantidad = cantidad
		}
	}

	if mejorCantidad*3 <= len(snapshots)*2 {
		log.Printf("Recuperacion rechazada: mejor consenso valido %d/%d respuestas recibidas", mejorCantidad, len(snapshots))
		return fmt.Errorf(
			"no hay consenso de inventario: %d/%d respuestas iguales",
			mejorCantidad,
			len(snapshots),
		)
	}

	estado.AplicarSnapshot(representante[mejorClave])
	log.Printf("Recuperacion aceptada: consenso valido %d/%d respuestas recibidas inventario=%s", mejorCantidad, len(snapshots), mejorClave)
	return nil
}
