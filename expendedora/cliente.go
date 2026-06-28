package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

var httpClient = &http.Client{
	Timeout: 500 * time.Millisecond,
}

// ParsearPeers transforma la lista de peers separada por comas en un slice.
func ParsearPeers(peersStr string) []string {
	if peersStr == "" {
		return nil
	}

	partes := strings.Split(peersStr, ",")

	var peers []string

	for _, p := range partes {
		p = strings.TrimSpace(p)

		if p != "" {
			peers = append(peers, p)
		}
	}

	return peers
}

// EsperarPeers intenta contactar a cada peer durante un tiempo maximo.
func EsperarPeers(peers []string, timeoutSeg int) {

	for _, peer := range peers {

		log.Printf("Esperando peer %s...", peer)

		deadline := time.Now().Add(
			time.Duration(timeoutSeg) * time.Second,
		)

		encontrado := false

		for time.Now().Before(deadline) {

			if PingPeer(peer) {
				log.Printf("Peer %s listo", peer)
				encontrado = true
				break
			}

			time.Sleep(200 * time.Millisecond)
		}

		if !encontrado {
			log.Printf(
				"Peer %s no respondió en %d segundos",
				peer,
				timeoutSeg,
			)
		}
	}
}

// EsperarPeersParalelo contacta todos los peers al mismo tiempo para no retrasar las instrucciones.
func EsperarPeersParalelo(peers []string, timeoutSeg int) {
	var wg sync.WaitGroup

	for _, peer := range peers {
		wg.Add(1)
		go func(peer string) {
			defer wg.Done()

			log.Printf("Esperando peer %s...", peer)

			deadline := time.Now().Add(
				time.Duration(timeoutSeg) * time.Second,
			)

			for time.Now().Before(deadline) {
				if PingPeer(peer) {
					log.Printf("Peer %s listo", peer)
					return
				}
				time.Sleep(200 * time.Millisecond)
			}

			log.Printf("Peer %s no respondio en %d segundos", peer, timeoutSeg)
		}(peer)
	}

	wg.Wait()
}

// PingPeer verifica si un peer REST esta disponible.
func PingPeer(peer string) bool {

	url := fmt.Sprintf(
		"http://%s/ping",
		peer,
	)

	resp, err := httpClient.Get(url)

	if err != nil {
		return false
	}

	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// EnviarInventario replica el inventario a un peer, o uno corrupto si esta infectado.
func EnviarInventario(
	peer string,
	inventario []Item,
	malicioso bool,
) error {

	payload := inventario
	log.Printf("Enviando inventario a %s: %v", peer, inventario)

	if malicioso {

		payload = []Item{
			{
				Nombre:   "CORRUPTO",
				Cantidad: 999999,
			},
		}

		log.Printf(
			"Enviando inventario CORRUPTO a %s",
			peer,
		)
	}

	data, err := json.Marshal(payload)

	if err != nil {
		return err
	}

	resp, err := httpClient.Post(
		fmt.Sprintf(
			"http://%s/inventario",
			peer,
		),
		"application/json",
		bytes.NewReader(data),
	)

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	log.Printf("Inventario enviado a %s con status %d", peer, resp.StatusCode)

	return nil
}

// EnviarVetos replica la lista de vetos a un peer.
func EnviarVetos(
	peer string,
	vetos map[string]int,
) error {
	log.Printf("Enviando vetos a %s: %v", peer, vetos)

	data, err := json.Marshal(vetos)

	if err != nil {
		return err
	}

	resp, err := httpClient.Post(
		fmt.Sprintf(
			"http://%s/vetos",
			peer,
		),
		"application/json",
		bytes.NewReader(data),
	)

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	log.Printf("Vetos enviados a %s con status %d", peer, resp.StatusCode)

	return nil
}


// BroadcastInventario envia el inventario a todos los peers concurrentemente.
func BroadcastInventario(
	peers []string,
	inventario []Item,
	malicioso bool,
) {
	log.Printf("Broadcast de inventario a %d peers", len(peers))

	for _, peer := range peers {

		go func(p string) {

			if err := EnviarInventario(
				p,
				inventario,
				malicioso,
			); err != nil {

				log.Printf(
					"Error enviando inventario a %s: %v",
					p,
					err,
				)
			}

		}(peer)
	}
}

// BroadcastVetos envia la lista de vetos a todos los peers concurrentemente.
func BroadcastVetos(
	peers []string,
	vetos map[string]int,
) {
	log.Printf("Broadcast de vetos a %d peers", len(peers))

	for _, peer := range peers {

		go func(p string) {

			if err := EnviarVetos(
				p,
				vetos,
			); err != nil {

				log.Printf(
					"Error enviando vetos a %s: %v",
					p,
					err,
				)
			}

		}(peer)
	}
}

// ObtenerSnapshot solicita a un peer su estado para recuperacion.
func ObtenerSnapshot(peer string) (SnapshotEstado, error) {
	log.Printf("Solicitando snapshot a %s", peer)
	resp, err := httpClient.Get(
		fmt.Sprintf(
			"http://%s/snapshot",
			peer,
		),
	)
	if err != nil {
		return SnapshotEstado{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return SnapshotEstado{}, fmt.Errorf("estado %d: %s", resp.StatusCode, string(body))
	}

	var snapshot SnapshotEstado
	if err := json.NewDecoder(resp.Body).Decode(&snapshot); err != nil {
		return SnapshotEstado{}, err
	}

	log.Printf("Snapshot recibido desde %s: inventario=%v vetos=%v malicioso=%v", peer, snapshot.Inventario, snapshot.Vetos, snapshot.Malicioso)
	return snapshot, nil
}

// BroadcastSnapshot sincroniza inventario y vetos con todos los peers.
func BroadcastSnapshot(peers []string, estado *Estado) {
	snapshot := estado.Snapshot()
	log.Printf("Sincronizacion periodica: inventario=%v vetos=%v peers=%d", snapshot.Inventario, snapshot.Vetos, len(peers))
	for _, peer := range peers {
		go func(p string) {
			EnviarInventario(p, snapshot.Inventario, estado.EsMalicioso())
			EnviarVetos(p, snapshot.Vetos)
		}(peer)
	}
}
