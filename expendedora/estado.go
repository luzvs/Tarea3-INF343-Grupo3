package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
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

func NuevoEstado(numMaquina, idProceso int) *Estado {
	return &Estado{
		NumMaquina: numMaquina,
		IdProceso:  idProceso,
		Vetos:      make(map[string]int),
	}
}

func (e *Estado) CargarInventarioAleatorio(carpeta string) error {
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

func (e *Estado) GetInventario() []Item {
	e.mu.Lock()
	defer e.mu.Unlock()
	copia := make([]Item, len(e.Inventario))
	copy(copia, e.Inventario)
	return copia
}

func (e *Estado) SetInventario(items []Item) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Inventario = items
}

func (e *Estado) GetVetos() map[string]int {
	e.mu.Lock()
	defer e.mu.Unlock()
	copia := make(map[string]int)
	for k, v := range e.Vetos {
		copia[k] = v
	}
	return copia
}

func (e *Estado) SetVetos(vetos map[string]int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Vetos = vetos
}

func (e *Estado) EstaVetado(nombre string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	counter, ok := e.Vetos[nombre]
	return ok && counter > 0
}

func (e *Estado) ToggleMalicioso() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Malicioso = !e.Malicioso
}

func (e *Estado) EsMalicioso() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Malicioso
}