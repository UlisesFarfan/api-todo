package handlers

import (
	"cmp"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	mathrand "math/rand/v2"
	"net/http"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"api-todo/internal/base"
	"api-todo/internal/httputil"
	"api-todo/internal/middleware"
)

// Un tablero (workspace en el front) es una sola fila, T#<id>/TABLERO, con sus
// columnas y notas adentro: el front siempre pide el tablero entero. Cada
// miembro tiene además una fila U#<email>/T#<id> para listar los suyos.
// shortcut: una fila de DynamoDB llega hasta 400 KB (unas 300 notas largas); si un tablero no entra, pasar columnas y notas a filas propias.
type tablero struct {
	Nombre string `json:"nombre"`
	// Miembros: email → rol. Por ahora solo el dueño ("owner").
	Miembros    map[string]string `json:"miembros"`
	Columnas    []columna         `json:"columnas"`
	Creado      time.Time         `json:"creado"`
	Actualizado time.Time         `json:"actualizado"`
}

// columna y nota se guardan y salen hacia el front con los mismos nombres.
type columna struct {
	ID          string    `json:"_id"`
	Nombre      string    `json:"name"`
	Notas       []nota    `json:"notes"`
	Creado      time.Time `json:"created_at"`
	Actualizado time.Time `json:"update_at"`
}

type nota struct {
	ID          string    `json:"_id"`
	Tarea       string    `json:"task"`
	Creado      time.Time `json:"created_at"`
	Actualizado time.Time `json:"update_at"`
}

type tableroJSON struct {
	ID     string `json:"_id"`
	Nombre string `json:"name"`
	Rol    string `json:"rol"`
}

type detalleJSON struct {
	tableroJSON
	Columnas []columna `json:"columns"`
}

const (
	maxTablero = 350 << 10
	maxNombre  = 100
	maxTarea   = 1000
)

var (
	errTableroNoEncontrado = errHTTP{http.StatusNotFound, "No encontramos ese tablero. Puede que lo hayan borrado."}
	errColumnaNoEncontrada = errHTTP{http.StatusNotFound, "No encontramos esa columna. Recargá la página."}
	errNotaNoEncontrada    = errHTTP{http.StatusNotFound, "No encontramos esa nota. Recargá la página."}
)

func pkTablero(id string) string { return "T#" + id }

// Las columnas y notas llevan adelante el id de su tablero ("<tablero>-<id>"):
// el front las edita sin decir de qué tablero son.
func nuevoID(tableroID string) string { return tableroID + "-" + rand.Text() }

func tableroDe(id string) string { t, _, _ := strings.Cut(id, "-"); return t }

// leerTablero devuelve el tablero si el usuario es miembro; si no, para él no existe.
func (h *Handlers) leerTablero(ctx context.Context, id, email string) (tablero, base.Fila, error) {
	var t tablero
	fila, existe, err := h.base.Leer(ctx, pkTablero(id), "TABLERO")
	if err == nil && existe {
		err = json.Unmarshal(fila.Datos, &t)
	}
	if err == nil && t.Miembros[email] == "" {
		err = errTableroNoEncontrado
	}
	return t, fila, err
}

// editar aplica f al tablero y lo guarda. Si otro pedido lo cambió en el medio
// (el front manda dos a la vez al mover una nota de columna), vuelve a leer y reintenta.
func (h *Handlers) editar(ctx context.Context, id, email string, f func(*tablero) error) error {
	for intento := range 8 {
		if intento > 0 {
			// espera al azar para no volver a chocar con el otro pedido
			time.Sleep(time.Duration(mathrand.IntN(20*intento)) * time.Millisecond)
		}
		t, fila, err := h.leerTablero(ctx, id, email)
		if err != nil {
			return err
		}
		if err := f(&t); err != nil {
			return err
		}
		t.Actualizado = time.Now()
		if fila.Datos, err = json.Marshal(t); err != nil {
			return err
		}
		if len(fila.Datos) > maxTablero {
			return invalido("El tablero está lleno. Borrá notas o columnas que ya no uses.")
		}
		if err = h.base.Escribir(ctx, base.Cambio{Fila: fila}); !errors.Is(err, base.ErrConflicto) {
			return err
		}
	}
	return errHTTP{http.StatusConflict, "El tablero cambió mientras lo editabas. Recargá la página."}
}

// ordenar deja primero los items de ids, en ese orden, y después el resto como
// estaban: si el front se olvida de alguno, no se pierde.
func ordenar[T any](items []T, ids []string, id func(T) string) {
	pos := map[string]int{}
	for i, x := range ids {
		if _, ok := pos[x]; !ok {
			pos[x] = i
		}
	}
	slices.SortStableFunc(items, func(a, b T) int {
		pa, oka := pos[id(a)]
		pb, okb := pos[id(b)]
		if oka && okb {
			return cmp.Compare(pa, pb)
		}
		if oka == okb {
			return 0
		}
		if oka {
			return -1
		}
		return 1
	})
}

func idColumna(c columna) string { return c.ID }
func idNota(n nota) string       { return n.ID }

func buscarColumna(t *tablero, id string) (int, error) {
	i := slices.IndexFunc(t.Columnas, func(c columna) bool { return c.ID == id })
	if i < 0 {
		return 0, errColumnaNoEncontrada
	}
	return i, nil
}

func buscarNota(t *tablero, id string) (int, int, error) {
	for i, c := range t.Columnas {
		if j := slices.IndexFunc(c.Notas, func(n nota) bool { return n.ID == id }); j >= 0 {
			return i, j, nil
		}
	}
	return 0, 0, errNotaNoEncontrada
}

// El nombre del tablero se manda al apretar Enter: tiene que tener algo.
func nombreTablero(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" || utf8.RuneCountInString(s) > maxNombre {
		return "", invalido("Escribí el nombre del tablero (hasta 100 letras).")
	}
	return s, nil
}

// Columnas y notas se guardan mientras se escribe: pueden quedar vacías un momento.
func nombreColumna(s string) error {
	if utf8.RuneCountInString(s) > maxNombre {
		return invalido("El nombre de la columna puede tener hasta 100 letras.")
	}
	return nil
}

func tarea(s string) error {
	if utf8.RuneCountInString(s) > maxTarea {
		return invalido("La nota puede tener hasta 1000 letras.")
	}
	return nil
}

// ---------- Tableros ----------

// GET /api/work_space: los tableros del usuario, del más viejo al más nuevo.
func (h *Handlers) Tableros(w http.ResponseWriter, r *http.Request) {
	ctx, email := r.Context(), middleware.UsuarioDe(r.Context()).Email
	filas, err := h.base.Consultar(ctx, middleware.PKUsuario(email), "T#")
	if err != nil {
		httputil.Interno(w, err)
		return
	}
	type item struct {
		tableroJSON
		creado time.Time
	}
	items := []item{}
	for _, f := range filas {
		id := strings.TrimPrefix(f.SK, "T#")
		t, _, err := h.leerTablero(ctx, id, email)
		if err != nil {
			responderError(w, err)
			return
		}
		items = append(items, item{tableroJSON{id, t.Nombre, t.Miembros[email]}, t.Creado})
	}
	slices.SortFunc(items, func(a, b item) int { return a.creado.Compare(b.creado) })
	lista := make([]tableroJSON, len(items))
	for i, it := range items {
		lista[i] = it.tableroJSON
	}
	httputil.JSON(w, http.StatusOK, "", lista)
}

// GET /api/work_space/{id}: el tablero con sus columnas y notas.
func (h *Handlers) Tablero(w http.ResponseWriter, r *http.Request) {
	id, email := r.PathValue("id"), middleware.UsuarioDe(r.Context()).Email
	t, _, err := h.leerTablero(r.Context(), id, email)
	if err != nil {
		responderError(w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, "", detalleJSON{tableroJSON{id, t.Nombre, t.Miembros[email]}, t.Columnas})
}

// POST /api/work_space {name}: el usuario de la sesión queda como dueño (el
// user_id que manda el front no se usa).
func (h *Handlers) CrearTablero(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Nombre string `json:"name"`
	}
	if !httputil.LeerJSON(w, r, &body) {
		return
	}
	nombre, err := nombreTablero(body.Nombre)
	if err != nil {
		responderError(w, err)
		return
	}
	email, id, ahora := middleware.UsuarioDe(r.Context()).Email, rand.Text(), time.Now()
	datos, _ := json.Marshal(tablero{Nombre: nombre, Miembros: map[string]string{email: "owner"}, Columnas: []columna{}, Creado: ahora, Actualizado: ahora})
	err = h.base.Escribir(r.Context(),
		base.Cambio{Fila: base.Fila{PK: pkTablero(id), SK: "TABLERO", Datos: datos}},
		base.Cambio{Fila: base.Fila{PK: middleware.PKUsuario(email), SK: pkTablero(id), Datos: []byte("{}")}})
	if err != nil {
		httputil.Interno(w, err)
		return
	}
	httputil.JSON(w, http.StatusCreated, "Tablero creado.", tableroJSON{id, nombre, "owner"})
}

// PATCH /api/work_space {_id, name}: le cambia el nombre.
func (h *Handlers) RenombrarTablero(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID     string `json:"_id"`
		Nombre string `json:"name"`
	}
	if !httputil.LeerJSON(w, r, &body) {
		return
	}
	nombre, err := nombreTablero(body.Nombre)
	if err == nil {
		err = h.editar(r.Context(), body.ID, middleware.UsuarioDe(r.Context()).Email, func(t *tablero) error {
			t.Nombre = nombre
			return nil
		})
	}
	if err != nil {
		responderError(w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, "Tablero guardado.", nil)
}

// PUT /api/work_space {_id, columns: [ids]}: el orden de las columnas.
func (h *Handlers) OrdenarColumnas(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID       string   `json:"_id"`
		Columnas []string `json:"columns"`
	}
	if !httputil.LeerJSON(w, r, &body) {
		return
	}
	err := h.editar(r.Context(), body.ID, middleware.UsuarioDe(r.Context()).Email, func(t *tablero) error {
		ordenar(t.Columnas, body.Columnas, idColumna)
		return nil
	})
	if err != nil {
		responderError(w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, "Columnas ordenadas.", nil)
}

// DELETE /api/work_space/{id}: borra el tablero con todo lo que tiene. Solo el dueño.
func (h *Handlers) BorrarTablero(w http.ResponseWriter, r *http.Request) {
	id, email := r.PathValue("id"), middleware.UsuarioDe(r.Context()).Email
	t, _, err := h.leerTablero(r.Context(), id, email)
	if err == nil && t.Miembros[email] != "owner" {
		err = errHTTP{http.StatusForbidden, "Solo el dueño puede borrar el tablero."}
	}
	if err != nil {
		responderError(w, err)
		return
	}
	cambios := []base.Cambio{{Fila: base.Fila{PK: pkTablero(id), SK: "TABLERO"}, Borrar: true}}
	for m := range t.Miembros {
		cambios = append(cambios, base.Cambio{Fila: base.Fila{PK: middleware.PKUsuario(m), SK: pkTablero(id)}, Borrar: true})
	}
	if err := h.base.Escribir(r.Context(), cambios...); err != nil {
		httputil.Interno(w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, "Tablero borrado.", nil)
}

// ---------- Columnas ----------

// POST /api/column {workspace_id, name}: agrega una columna vacía al final.
func (h *Handlers) CrearColumna(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Tablero string `json:"workspace_id"`
		Nombre  string `json:"name"`
	}
	if !httputil.LeerJSON(w, r, &body) {
		return
	}
	var col columna
	err := nombreColumna(body.Nombre)
	if err == nil {
		err = h.editar(r.Context(), body.Tablero, middleware.UsuarioDe(r.Context()).Email, func(t *tablero) error {
			ahora := time.Now()
			col = columna{ID: nuevoID(body.Tablero), Nombre: body.Nombre, Notas: []nota{}, Creado: ahora, Actualizado: ahora}
			t.Columnas = append(t.Columnas, col)
			return nil
		})
	}
	if err != nil {
		responderError(w, err)
		return
	}
	httputil.JSON(w, http.StatusCreated, "Columna creada.", col)
}

// PATCH /api/column {_id, name}: le cambia el nombre.
func (h *Handlers) RenombrarColumna(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID     string `json:"_id"`
		Nombre string `json:"name"`
	}
	if !httputil.LeerJSON(w, r, &body) {
		return
	}
	err := nombreColumna(body.Nombre)
	if err == nil {
		err = h.editar(r.Context(), tableroDe(body.ID), middleware.UsuarioDe(r.Context()).Email, func(t *tablero) error {
			i, err := buscarColumna(t, body.ID)
			if err == nil {
				t.Columnas[i].Nombre, t.Columnas[i].Actualizado = body.Nombre, time.Now()
			}
			return err
		})
	}
	if err != nil {
		responderError(w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, "Columna guardada.", nil)
}

// PUT /api/column {_id, notes: [ids]}: las notas de la columna, en orden. Las
// que estén en otra columna se mueven acá; las de acá que no vengan quedan al
// final: al mover una nota el front manda la columna de origen y la de destino
// en dos pedidos, en cualquier orden, y así no se pierde en el medio.
func (h *Handlers) OrdenarNotas(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID    string   `json:"_id"`
		Notas []string `json:"notes"`
	}
	if !httputil.LeerJSON(w, r, &body) {
		return
	}
	err := h.editar(r.Context(), tableroDe(body.ID), middleware.UsuarioDe(r.Context()).Email, func(t *tablero) error {
		i, err := buscarColumna(t, body.ID)
		if err != nil {
			return err
		}
		quiero := map[string]bool{}
		for _, id := range body.Notas {
			quiero[id] = true
		}
		for j := range t.Columnas {
			if j == i {
				continue
			}
			quedan := t.Columnas[j].Notas[:0]
			for _, n := range t.Columnas[j].Notas {
				if quiero[n.ID] {
					t.Columnas[i].Notas = append(t.Columnas[i].Notas, n)
				} else {
					quedan = append(quedan, n)
				}
			}
			t.Columnas[j].Notas = quedan
		}
		ordenar(t.Columnas[i].Notas, body.Notas, idNota)
		return nil
	})
	if err != nil {
		responderError(w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, "Notas ordenadas.", nil)
}

// DELETE /api/column/{id}: borra la columna con sus notas.
func (h *Handlers) BorrarColumna(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := h.editar(r.Context(), tableroDe(id), middleware.UsuarioDe(r.Context()).Email, func(t *tablero) error {
		i, err := buscarColumna(t, id)
		if err == nil {
			t.Columnas = slices.Delete(t.Columnas, i, i+1)
		}
		return err
	})
	if err != nil {
		responderError(w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, "Columna borrada.", nil)
}

// ---------- Notas ----------

// POST /api/note {column_id, task}: agrega la nota al final de la columna.
func (h *Handlers) CrearNota(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Columna string `json:"column_id"`
		Tarea   string `json:"task"`
	}
	if !httputil.LeerJSON(w, r, &body) {
		return
	}
	var n nota
	err := tarea(body.Tarea)
	if err == nil {
		err = h.editar(r.Context(), tableroDe(body.Columna), middleware.UsuarioDe(r.Context()).Email, func(t *tablero) error {
			i, err := buscarColumna(t, body.Columna)
			if err != nil {
				return err
			}
			ahora := time.Now()
			n = nota{ID: nuevoID(tableroDe(body.Columna)), Tarea: body.Tarea, Creado: ahora, Actualizado: ahora}
			t.Columnas[i].Notas = append(t.Columnas[i].Notas, n)
			return nil
		})
	}
	if err != nil {
		responderError(w, err)
		return
	}
	httputil.JSON(w, http.StatusCreated, "Nota creada.", n)
}

// PATCH /api/note {_id, task}: cambia el texto.
func (h *Handlers) EditarNota(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID    string `json:"_id"`
		Tarea string `json:"task"`
	}
	if !httputil.LeerJSON(w, r, &body) {
		return
	}
	err := tarea(body.Tarea)
	if err == nil {
		err = h.editar(r.Context(), tableroDe(body.ID), middleware.UsuarioDe(r.Context()).Email, func(t *tablero) error {
			i, j, err := buscarNota(t, body.ID)
			if err == nil {
				t.Columnas[i].Notas[j].Tarea, t.Columnas[i].Notas[j].Actualizado = body.Tarea, time.Now()
			}
			return err
		})
	}
	if err != nil {
		responderError(w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, "Nota guardada.", nil)
}

// DELETE /api/note/{id}
func (h *Handlers) BorrarNota(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := h.editar(r.Context(), tableroDe(id), middleware.UsuarioDe(r.Context()).Email, func(t *tablero) error {
		i, j, err := buscarNota(t, id)
		if err == nil {
			t.Columnas[i].Notas = slices.Delete(t.Columnas[i].Notas, j, j+1)
		}
		return err
	})
	if err != nil {
		responderError(w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, "Nota borrada.", nil)
}
