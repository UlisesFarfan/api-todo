// Package base guarda todo en una tabla de DynamoDB (o en memoria, en local y
// en los tests). Cada fila es una clave (pk, sk), sus datos en JSON y una
// versión: toda escritura dice en qué versión espera encontrar la fila, así dos
// pedidos a la vez sobre el mismo tablero no se pisan.
package base

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"
)

// ErrConflicto: alguna fila no estaba en la versión esperada. Se vuelve a leer y se reintenta.
var ErrConflicto = errors.New("la fila cambió mientras se editaba")

type Fila struct {
	PK, SK  string
	Datos   []byte
	Version int64
	// Vence (unix): DynamoDB borra la fila sola un rato después (TTL). 0: no vence.
	Vence int64
}

// Cambio guarda la fila si sigue en Fila.Version (0: si no existe) y la deja en
// Version+1. Con Borrar la borra; con Version 0 la borra sin mirar la versión.
type Cambio struct {
	Fila
	Borrar bool
}

type Base interface {
	// Leer devuelve la fila y si existe.
	Leer(ctx context.Context, pk, sk string) (Fila, bool, error)
	// Consultar devuelve las filas de pk cuyo sk empieza con prefijo, ordenadas por sk.
	Consultar(ctx context.Context, pk, prefijo string) ([]Fila, error)
	// Escribir aplica todos los cambios o ninguno (hasta 100); ErrConflicto si falla una versión.
	Escribir(ctx context.Context, cambios ...Cambio) error
}

// Memoria es la base de desarrollo y de los tests: se pierde al reiniciar.
type Memoria struct {
	mu    sync.Mutex
	filas map[[2]string]Fila
}

func NuevaMemoria() *Memoria { return &Memoria{filas: map[[2]string]Fila{}} }

func (m *Memoria) Leer(_ context.Context, pk, sk string) (Fila, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	f, ok := m.filas[[2]string{pk, sk}]
	f.Datos = slices.Clone(f.Datos)
	return f, ok, nil
}

func (m *Memoria) Consultar(_ context.Context, pk, prefijo string) ([]Fila, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var filas []Fila
	for k, f := range m.filas {
		if k[0] == pk && strings.HasPrefix(k[1], prefijo) {
			f.Datos = slices.Clone(f.Datos)
			filas = append(filas, f)
		}
	}
	slices.SortFunc(filas, func(a, b Fila) int { return strings.Compare(a.SK, b.SK) })
	return filas, nil
}

func (m *Memoria) Escribir(_ context.Context, cambios ...Cambio) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range cambios {
		actual, existe := m.filas[[2]string{c.PK, c.SK}]
		if c.Borrar && c.Version == 0 {
			continue
		}
		if (c.Version == 0 && existe) || (c.Version != 0 && (!existe || actual.Version != c.Version)) {
			return ErrConflicto
		}
	}
	for _, c := range cambios {
		k := [2]string{c.PK, c.SK}
		if c.Borrar {
			delete(m.filas, k)
			continue
		}
		f := c.Fila
		f.Datos = slices.Clone(f.Datos)
		f.Version++
		m.filas[k] = f
	}
	return nil
}
