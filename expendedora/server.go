package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Servidor struct {
	puerto string
	estado *Estado
	mux    *http.ServeMux
}

func NuevoServidor(puerto string, estado *Estado) *Servidor {
	s := &Servidor{
		puerto: puerto,
		estado: estado,
		mux:    http.NewServeMux(),
	}

	s.registrarRutas()

	return s
}

func (s *Servidor) Iniciar() error {
	log.Printf(
		"[M%dP%d] Escuchando en puerto %s",
		s.estado.NumMaquina,
		s.estado.IdProceso,
		s.puerto,
	)

	return http.ListenAndServe(":"+s.puerto, s.mux)
}

func (s *Servidor) registrarRutas() {

	s.mux.HandleFunc("/ping", s.handlePing)

	s.mux.HandleFunc("/inventario", s.handleInventario)

	s.mux.HandleFunc("/vetos", s.handleVetos)

	s.mux.HandleFunc("/estado", s.handleEstado)

	s.mux.HandleFunc("/infectar", s.handleInfectar)
}

func (s *Servidor) handlePing(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "pong")
}

func (s *Servidor) handleInventario(
	w http.ResponseWriter,
	r *http.Request,
) {

	if r.Method != http.MethodPost {
		http.Error(
			w,
			"metodo no permitido",
			http.StatusMethodNotAllowed,
		)
		return
	}

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(
			w,
			"error leyendo body",
			http.StatusBadRequest,
		)
		return
	}

	var items []Item

	if err := json.Unmarshal(body, &items); err != nil {
		http.Error(
			w,
			"json invalido",
			http.StatusBadRequest,
		)
		return
	}

	s.estado.SetInventario(items)

	log.Printf(
		"[M%dP%d] Inventario actualizado",
		s.estado.NumMaquina,
		s.estado.IdProceso,
	)

	w.WriteHeader(http.StatusOK)
}

func (s *Servidor) handleVetos(
	w http.ResponseWriter,
	r *http.Request,
) {

	if r.Method != http.MethodPost {
		http.Error(
			w,
			"metodo no permitido",
			http.StatusMethodNotAllowed,
		)
		return
	}

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(
			w,
			"error leyendo body",
			http.StatusBadRequest,
		)
		return
	}

	var vetos map[string]int

	if err := json.Unmarshal(body, &vetos); err != nil {
		http.Error(
			w,
			"json invalido",
			http.StatusBadRequest,
		)
		return
	}

	s.estado.SetVetos(vetos)

	log.Printf(
		"[M%dP%d] Vetos actualizados",
		s.estado.NumMaquina,
		s.estado.IdProceso,
	)

	w.WriteHeader(http.StatusOK)
}

func (s *Servidor) handleEstado(
	w http.ResponseWriter,
	r *http.Request,
) {

	if r.Method != http.MethodGet {
		http.Error(
			w,
			"metodo no permitido",
			http.StatusMethodNotAllowed,
		)
		return
	}

	resp := map[string]interface{}{
		"inventario": s.estado.GetInventario(),
		"vetos":      s.estado.GetVetos(),
		"malicioso":  s.estado.EsMalicioso(),
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(resp)
}

func (s *Servidor) handleInfectar(
	w http.ResponseWriter,
	r *http.Request,
) {

	if r.Method != http.MethodPost {
		http.Error(
			w,
			"metodo no permitido",
			http.StatusMethodNotAllowed,
		)
		return
	}

	s.estado.ToggleMalicioso()

	w.WriteHeader(http.StatusOK)
}