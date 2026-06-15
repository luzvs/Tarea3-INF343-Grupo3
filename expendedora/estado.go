package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type Item struct {
	Nombre   string `json:"nombre"`
	Cantidad int    `json:"cantidad"`
}

type Estado struct {
	mu         sync.Mutex
	NumMaquina int
	IdProceso  int
	Inventario []Item
	Vetos      map[string]int
	Malicioso  bool
}

// NuevoEstado crea el estado inicial protegido por mutex para un proceso.
func NuevoEstado(numMaquina, idProceso int) *Estado {
	return &Estado{
		NumMaquina: numMaquina,
		IdProceso:  idProceso,
		Vetos:      make(map[string]int),
	}
}

// CargarInventarioAleatorio selecciona y carga un inventario JSON desde disco.
func (e *Estado) CargarInventarioAleatorio(carpeta string) error {
	rand.Seed(time.Now().UnixNano() + int64(e.NumMaquina*1000+e.IdProceso))
	archivos, err := filepath.Glob(filepath.Join(carpeta, "*.json"))
	if err != nil || len(archivos) == 0 {
		return fmt.Errorf("no se encontraron inventarios en %s", carpeta)
	}
	elegido := archivos[rand.Intn(len(archivos))]
	data, err := os.ReadFile(elegido)
	if err != nil {
		return fmt.Errorf("error leyendo %s: %v", elegido, err)
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	return json.Unmarshal(data, &e.Inventario)
}

// GetInventario retorna una copia del inventario actual.
func (e *Estado) GetInventario() []Item {
	e.mu.Lock()
	defer e.mu.Unlock()
	copia := make([]Item, len(e.Inventario))
	copy(copia, e.Inventario)
	return copia
}

// SetInventario reemplaza el inventario local con una copia recibida.
func (e *Estado) SetInventario(items []Item) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Inventario = copiarInventario(items)
}

// GetVetos retorna una copia de la lista de vetos.
func (e *Estado) GetVetos() map[string]int {
	e.mu.Lock()
	defer e.mu.Unlock()
	copia := make(map[string]int)
	for k, v := range e.Vetos {
		copia[k] = v
	}
	return copia
}

// SetVetos reemplaza la lista local de vetos con una copia recibida.
func (e *Estado) SetVetos(vetos map[string]int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Vetos = copiarVetos(vetos)
}

// EstaVetado indica si una persona tiene un counter de veto activo.
func (e *Estado) EstaVetado(nombre string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	counter, ok := e.Vetos[nombre]
	return ok && counter > 0
}

// ToggleMalicioso alterna el modo de envio corrupto de inventarios.
func (e *Estado) ToggleMalicioso() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Malicioso = !e.Malicioso
}

// EsMalicioso retorna si el proceso esta infectado.
func (e *Estado) EsMalicioso() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Malicioso
}

// Vetar agrega o reinicia el veto de una persona con counter 5.
func (e *Estado) Vetar(nombre string) map[string]int {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Vetos[nombre] = 5
	return copiarVetos(e.Vetos)
}

// Perdonar elimina una persona de la lista de vetos.
func (e *Estado) Perdonar(nombre string) map[string]int {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.Vetos, nombre)
	return copiarVetos(e.Vetos)
}

// Comprar valida veto y stock, descuenta inventario y retorna el resultado.
func (e *Estado) Comprar(nombre, producto string, cantidad int) string {
	e.mu.Lock()
	defer e.mu.Unlock()

	if counter, ok := e.Vetos[nombre]; ok && counter > 0 {
		e.decrementarVetosBloqueado()
		return "DENEGADO"
	}

	for i := range e.Inventario {
		if e.Inventario[i].Nombre == producto {
			if e.Inventario[i].Cantidad < cantidad {
				e.decrementarVetosBloqueado()
				return "NO VALIDO"
			}
			e.Inventario[i].Cantidad -= cantidad
			e.decrementarVetosBloqueado()
			return "VALIDO"
		}
	}

	e.decrementarVetosBloqueado()
	return "NO VALIDO"
}

// DecrementarVetos reduce los counters activos y elimina los vencidos.
func (e *Estado) DecrementarVetos() map[string]int {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.decrementarVetosBloqueado()
	return copiarVetos(e.Vetos)
}

// Snapshot construye una copia serializable del estado local.
func (e *Estado) Snapshot() SnapshotEstado {
	e.mu.Lock()
	defer e.mu.Unlock()
	return SnapshotEstado{
		Inventario: copiarInventario(e.Inventario),
		Vetos:      copiarVetos(e.Vetos),
		Malicioso:  e.Malicioso,
	}
}

// AplicarSnapshot reemplaza el estado local con un snapshot recuperado.
func (e *Estado) AplicarSnapshot(s SnapshotEstado) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Inventario = copiarInventario(s.Inventario)
	e.Vetos = copiarVetos(s.Vetos)
	e.Malicioso = s.Malicioso
}

// decrementarVetosBloqueado actualiza counters asumiendo que el mutex esta tomado.
func (e *Estado) decrementarVetosBloqueado() {
	for nombre, counter := range e.Vetos {
		counter--
		if counter <= 0 {
			delete(e.Vetos, nombre)
		} else {
			e.Vetos[nombre] = counter
		}
	}
}

// copiarInventario evita compartir slices mutables entre procesos internos.
func copiarInventario(items []Item) []Item {
	copia := make([]Item, len(items))
	copy(copia, items)
	return copia
}

// copiarVetos evita compartir mapas mutables entre procesos internos.
func copiarVetos(vetos map[string]int) map[string]int {
	copia := make(map[string]int, len(vetos))
	for k, v := range vetos {
		copia[k] = v
	}
	return copia
}

// ClaveInventario normaliza un inventario para comparar replicas.
func ClaveInventario(items []Item) string {
	copia := copiarInventario(items)
	sort.Slice(copia, func(i, j int) bool {
		return copia[i].Nombre < copia[j].Nombre
	})
	data, _ := json.Marshal(copia)
	return string(data)
}

type SnapshotEstado struct {
	Inventario []Item         `json:"inventario"`
	Vetos      map[string]int `json:"vetos"`
	Malicioso  bool           `json:"malicioso"`
}
