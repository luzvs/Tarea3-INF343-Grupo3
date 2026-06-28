package main

import (
	"fmt"
	"log"
	"time"
)

// RecuperarEstado solicita snapshots y recupera usando mayoria de inventarios Y vetos.
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

// aplicarMayoritaria valida que mas de dos tercios compartan el mismo inventario
// y luego aplica quorum independiente sobre los vetos para garantizar convergencia
// determinista en ambas dimensiones del estado.
func aplicarMayoritaria(estado *Estado, snapshots []SnapshotEstado) error {
	if len(snapshots) == 0 {
		return fmt.Errorf("no se recibieron estados para recuperar")
	}

	// --- Quorum de inventario ---
	conteoInv := make(map[string]int)
	representante := make(map[string]SnapshotEstado)

	for _, snapshot := range snapshots {
		if EsInventarioCorrupto(snapshot.Inventario) {
			log.Printf("Recuperacion: snapshot corrupto recibido; cuenta contra el consenso")
			continue
		}

		clave := ClaveInventario(snapshot.Inventario)
		conteoInv[clave]++
		representante[clave] = snapshot
		log.Printf("Recuperacion: inventario candidato=%s conteo=%d", clave, conteoInv[clave])
	}

	if len(conteo) == 0 {
		return fmt.Errorf("no se recibieron inventarios validos para recuperar")
	}

	var mejorClave string
	var mejorCantidad int
	for clave, cantidad := range conteoInv {
		if cantidad > mejorCantidad {
			mejorClave = clave
			mejorCantidad = cantidad
		}
	}

	// Condicion correcta: necesitamos ESTRICTAMENTE mas de 2/3.
	// "mejorCantidad > (2/3)*len(snapshots)"  equivale a  "mejorCantidad*3 > len(snapshots)*2"
	if mejorCantidad*3 <= len(snapshots)*2 {
		log.Printf("Recuperacion rechazada: mejor consenso valido %d/%d respuestas recibidas", mejorCantidad, len(snapshots))
		return fmt.Errorf(
			"no hay consenso de inventario: %d/%d respuestas iguales",
			mejorCantidad,
			len(snapshots),
		)
	}

<<<<<<< HEAD
	estado.AplicarSnapshot(representante[mejorClave])
	log.Printf("Recuperacion aceptada: consenso valido %d/%d respuestas recibidas inventario=%s", mejorCantidad, len(snapshots), mejorClave)
=======
	// --- Quorum de vetos (independiente del inventario) ---
	// Serializamos cada mapa de vetos como clave comparable.
	conteoVetos := make(map[string]int)
	representanteVetos := make(map[string]map[string]int)

	for _, snapshot := range snapshots {
		clave := ClaveVetos(snapshot.Vetos)
		conteoVetos[clave]++
		representanteVetos[clave] = snapshot.Vetos
	}

	var mejorClaveVetos string
	var mejorCantidadVetos int
	for clave, cantidad := range conteoVetos {
		if cantidad > mejorCantidadVetos {
			mejorClaveVetos = clave
			mejorCantidadVetos = cantidad
		}
	}

	// Si no hay quorum en vetos tampoco hay convergencia determinista.
	// Usamos la misma regla estricta de 2/3.
	if mejorCantidadVetos*3 <= len(snapshots)*2 {
		log.Printf("Recuperacion rechazada por vetos: mejor consenso %d/%d", mejorCantidadVetos, len(snapshots))
		return fmt.Errorf(
			"no hay consenso de vetos: %d/%d respuestas iguales",
			mejorCantidadVetos,
			len(snapshots),
		)
	}

	// Aplicar inventario con quorum + vetos con quorum
	snapshotFinal := representante[mejorClave]
	snapshotFinal.Vetos = representanteVetos[mejorClaveVetos]
	estado.AplicarSnapshot(snapshotFinal)

	log.Printf(
		"Recuperacion aceptada: inventario consenso=%d/%d vetos consenso=%d/%d",
		mejorCantidad, len(snapshots),
		mejorCantidadVetos, len(snapshots),
	)
>>>>>>> a57af2d96c3bbd14dca9bbef51410491bc0c0f5b
	return nil
}
