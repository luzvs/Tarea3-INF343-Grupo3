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

// NuevoServidor crea un servidor HTTP asociado al estado del proceso.
func NuevoServidor(puerto string, estado *Estado) *Servidor {
	s := &Servidor{
		puerto: puerto,
		estado: estado,
		mux:    http.NewServeMux(),
	}

	s.registrarRutas()

	return s
}

// Iniciar deja el servidor escuchando en el puerto configurado.
func (s *Servidor) Iniciar() error {
	log.Printf(
		"[M%dP%d] Escuchando en puerto %s",
		s.estado.NumMaquina,
		s.estado.IdProceso,
		s.puerto,
	)

	return http.ListenAndServe(":"+s.puerto, s.mux)
}

// registrarRutas asocia los endpoints REST con sus handlers.
func (s *Servidor) registrarRutas() {

	s.mux.HandleFunc("/ping", s.handlePing)

	s.mux.HandleFunc("/inventario", s.handleInventario)

	s.mux.HandleFunc("/vetos", s.handleVetos)

	s.mux.HandleFunc("/estado", s.handleEstado)

	s.mux.HandleFunc("/infectar", s.handleInfectar)

	s.mux.HandleFunc("/snapshot", s.handleSnapshot)
}

// handlePing responde disponibilidad basica del proceso.
func (s *Servidor) handlePing(
	w http.ResponseWriter,
	r *http.Request,
) {
	log.Printf(
		"[M%dP%d] Recibido /ping desde %s",
		s.estado.NumMaquina,
		s.estado.IdProceso,
		r.RemoteAddr,
	)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "pong")
}

// handleInventario recibe una replica de inventario desde otro proceso.
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

	if EsInventarioCorrupto(items) {
		log.Printf(
			"[M%dP%d] Inventario corrupto rechazado desde %s",
			s.estado.NumMaquina,
			s.estado.IdProceso,
			r.RemoteAddr,
		)
		http.Error(
			w,
			"inventario corrupto rechazado",
			http.StatusConflict,
		)
		return
	}

	log.Printf(
		"[M%dP%d] Recibido inventario desde %s: %v",
		s.estado.NumMaquina,
		s.estado.IdProceso,
		r.RemoteAddr,
		items,
	)

	s.estado.SetInventario(items)

	log.Printf(
		"[M%dP%d] Inventario actualizado",
		s.estado.NumMaquina,
		s.estado.IdProceso,
	)

	w.WriteHeader(http.StatusOK)
}

// handleVetos recibe una replica de la lista de vetos.
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

	log.Printf(
		"[M%dP%d] Recibidos vetos desde %s: %v",
		s.estado.NumMaquina,
		s.estado.IdProceso,
		r.RemoteAddr,
		vetos,
	)

	s.estado.SetVetos(vetos)

	log.Printf(
		"[M%dP%d] Vetos actualizados",
		s.estado.NumMaquina,
		s.estado.IdProceso,
	)

	w.WriteHeader(http.StatusOK)
}

// handleEstado expone inventario, vetos y modo malicioso para inspeccion.
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
	log.Printf(
		"[M%dP%d] Consulta de estado desde %s: %v",
		s.estado.NumMaquina,
		s.estado.IdProceso,
		r.RemoteAddr,
		resp,
	)

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(resp)
}

// handleInfectar alterna el modo malicioso del proceso.
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
	log.Printf(
		"[M%dP%d] Modo malicioso cambiado a %v por %s",
		s.estado.NumMaquina,
		s.estado.IdProceso,
		s.estado.EsMalicioso(),
		r.RemoteAddr,
	)

	w.WriteHeader(http.StatusOK)
}

// handleSnapshot entrega el estado usado por la recuperacion.
func (s *Servidor) handleSnapshot(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(
			w,
			"metodo no permitido",
			http.StatusMethodNotAllowed,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	snapshot := s.estado.Snapshot()
	if snapshot.Malicioso {
		snapshot.Inventario = []Item{{Nombre: "CORRUPTO", Cantidad: 999999}}
	}
	log.Printf(
		"[M%dP%d] Enviando snapshot a %s: inventario=%v vetos=%v malicioso=%v",
		s.estado.NumMaquina,
		s.estado.IdProceso,
		r.RemoteAddr,
		snapshot.Inventario,
		snapshot.Vetos,
		snapshot.Malicioso,
	)
	json.NewEncoder(w).Encode(snapshot)
}
