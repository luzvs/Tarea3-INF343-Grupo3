
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Ejecutor struct {
	estado     *Estado
	peers      []string
	numMaquina int
	idProceso  int
}

// NuevoEjecutor crea el lector de instrucciones asociado a un proceso.
func NuevoEjecutor(estado *Estado, peers []string, numMaquina, idProceso int) *Ejecutor {
	return &Ejecutor{
		estado:     estado,
		peers:      peers,
		numMaquina: numMaquina,
		idProceso:  idProceso,
	}
}

// BuscarArchivo encuentra el archivo de instrucciones que termina en _ID.txt.
func BuscarArchivo(carpeta string, id int) string {
	patron := filepath.Join(carpeta, fmt.Sprintf("*_%d.txt", id))
	matches, err := filepath.Glob(patron)
	if err != nil || len(matches) == 0 {
		return ""
	}
	return matches[0]
}

// EjecutarArchivo lee y ejecuta secuencialmente las instrucciones del proceso.
func (e *Ejecutor) EjecutarArchivo(archivo string) error {
	f, err := os.Open(archivo)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		linea := strings.TrimSpace(scanner.Text())
		if linea == "" || strings.HasPrefix(linea, "#") {
			continue
		}
		log.Printf("[M%dP%d] Instruccion leida: %s", e.numMaquina, e.idProceso, linea)
		e.ejecutarLinea(linea)
	}

	return scanner.Err()
}

// ejecutarLinea aplica una instruccion VETAR, COMPRAR o PERDONAR.
func (e *Ejecutor) ejecutarLinea(linea string) {
	partes := strings.Fields(linea)
	if len(partes) == 0 {
		return
	}

	switch partes[0] {
	case "VETAR":
		if len(partes) < 2 {
			e.escribirLogInventario(linea, "NO VALIDO")
			return
		}
		nombre := strings.Join(partes[1:], " ")
		log.Printf("[M%dP%d] Ejecutando VETAR sobre %q", e.numMaquina, e.idProceso, nombre)
		vetos := e.estado.Vetar(nombre)
		e.escribirLogInventario(linea, "")
		e.escribirLogVetos(vetos)
		log.Printf("[M%dP%d] Veto aplicado: vetos=%v", e.numMaquina, e.idProceso, vetos)
		BroadcastVetos(e.peers, vetos)

	case "PERDONAR":
		if len(partes) < 2 {
			e.escribirLogInventario(linea, "NO VALIDO")
			return
		}
		nombre := strings.Join(partes[1:], " ")
		log.Printf("[M%dP%d] Ejecutando PERDONAR sobre %q", e.numMaquina, e.idProceso, nombre)
		vetos := e.estado.Perdonar(nombre)
		e.escribirLogInventario(linea, "")
		e.escribirLogVetos(vetos)
		log.Printf("[M%dP%d] Perdon aplicado: vetos=%v", e.numMaquina, e.idProceso, vetos)
		BroadcastVetos(e.peers, vetos)

	case "COMPRAR":
		nombre, producto, cantidad, ok := parsearCompra(partes)
		if !ok {
			e.escribirLogInventario(linea, "NO VALIDO")
			return
		}

		resultado := e.estado.Comprar(nombre, producto, cantidad)
		e.escribirLogInventario(linea, resultado)
		e.escribirLogVetos(e.estado.GetVetos())
		log.Printf(
			"[M%dP%d] Compra persona=%q producto=%q cantidad=%d resultado=%s inventario=%v vetos=%v",
			e.numMaquina,
			e.idProceso,
			nombre,
			producto,
			cantidad,
			resultado,
			e.estado.GetInventario(),
			e.estado.GetVetos(),
		)

		if resultado == "VALIDO" {
			BroadcastInventario(e.peers, e.estado.GetInventario(), e.estado.EsMalicioso())
		}
		BroadcastVetos(e.peers, e.estado.GetVetos())

	default:
		e.escribirLogInventario(linea, "NO VALIDO")
	}
}

// parsearCompra separa nombre, producto y cantidad permitiendo nombres con espacios.
func parsearCompra(partes []string) (string, string, int, bool) {
	if len(partes) < 4 {
		return "", "", 0, false
	}

	cantidad, err := strconv.Atoi(partes[len(partes)-1])
	if err != nil || cantidad <= 0 {
		return "", "", 0, false
	}

	producto := partes[len(partes)-2]
	nombre := strings.Join(partes[1:len(partes)-2], " ")
	if nombre == "" || producto == "" {
		return "", "", 0, false
	}

	return nombre, producto, cantidad, true
}

// escribirLogInventario agrega una linea al log de inventario del proceso.
func (e *Ejecutor) escribirLogInventario(instruccion, resultado string) {
	logPath := filepath.Join(
		"logs",
		fmt.Sprintf("inventario_M%dP%d.log", e.numMaquina, e.idProceso),
	)
	escribirLineaLog(logPath, formatearLinea(instruccion, resultado))
}

// escribirLogVetos reescribe el log de vetos con el estado actual.
func (e *Ejecutor) escribirLogVetos(vetos map[string]int) {
	logPath := filepath.Join(
		"logs",
		fmt.Sprintf("vetos_M%dP%d.log", e.numMaquina, e.idProceso),
	)

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	for nombre, counter := range vetos {
		fmt.Fprintf(f, "VETADO %s %d\n", nombre, counter)
	}
}

// formatearLinea construye la linea de log con resultado cuando corresponde.
func formatearLinea(instruccion, resultado string) string {
	if resultado == "" {
		return instruccion
	}
	return fmt.Sprintf("%s | %s", instruccion, resultado)
}

// escribirLineaLog crea el archivo si falta y agrega una linea.
func escribirLineaLog(logPath, linea string) {
	_ = os.MkdirAll(filepath.Dir(logPath), 0755)
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintln(f, linea)
}

// escribirLog mantiene compatibilidad con el nombre usado durante pruebas locales.
func (e *Ejecutor) escribirLog(instruccion, resultado string) {
	e.escribirLogInventario(instruccion, resultado)
}
