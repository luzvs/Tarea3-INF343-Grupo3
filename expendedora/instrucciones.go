
//Son funciones falsas que use para probarlo local
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type Ejecutor struct {
	estado     *Estado
	peers      []string
	numMaquina int
	idProceso  int
}

func NuevoEjecutor(estado *Estado, peers []string, numMaquina, idProceso int) *Ejecutor {
	return &Ejecutor{
		estado:     estado,
		peers:      peers,
		numMaquina: numMaquina,
		idProceso:  idProceso,
	}
}

func BuscarArchivo(carpeta string, id int) string {
	patron := filepath.Join(carpeta, fmt.Sprintf("*_%d.txt", id))
	matches, err := filepath.Glob(patron)
	if err != nil || len(matches) == 0 {
		return ""
	}
	return matches[0]
}

func (e *Ejecutor) EjecutarArchivo(archivo string) error {
	return nil
}

func (e *Ejecutor) escribirLog(instruccion, resultado string) {
	logPath := fmt.Sprintf("../logs/instrucciones_M%dP%d.log", e.numMaquina, e.idProceso)
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	linea := instruccion
	if resultado != "" {
		linea = fmt.Sprintf("%s | %s", instruccion, resultado)
	}
	fmt.Fprintln(f, linea)
}