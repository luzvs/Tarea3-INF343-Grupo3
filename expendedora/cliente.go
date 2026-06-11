package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

var httpClient = &http.Client{
	Timeout: 500 * time.Millisecond,
}

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

func EnviarInventario(
	peer string,
	inventario []Item,
	malicioso bool,
) error {

	payload := inventario

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

	return nil
}

func EnviarVetos(
	peer string,
	vetos map[string]int,
) error {

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

	return nil
}


func BroadcastInventario(
	peers []string,
	inventario []Item,
	malicioso bool,
) {

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

func BroadcastVetos(
	peers []string,
	vetos map[string]int,
) {

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