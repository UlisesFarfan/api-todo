package router

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"api-todo/internal/base"
	"api-todo/internal/handlers"
)

// La API entera contra una base en memoria, usada por HTTP como la usa el front.
// El login con Google es falso: el credential "google|<sub>|<email>" es válido.
func nuevaAPI(t *testing.T) *httptest.Server {
	srv := httptest.NewServer(New(base.NuevaMemoria(), []string{"https://front.test"}, googleFalso))
	t.Cleanup(srv.Close)
	return srv
}

func googleFalso(_ context.Context, credential string) (handlers.PerfilGoogle, error) {
	partes := strings.Split(credential, "|")
	if len(partes) != 3 || partes[0] != "google" {
		return handlers.PerfilGoogle{}, errors.New("token inválido")
	}
	return handlers.PerfilGoogle{Sub: partes[1], Email: partes[2], Nombre: "Ana", Foto: "https://foto/ana.png"}, nil
}

type resp struct {
	status int
	Code   int             `json:"code"`
	Msg    string          `json:"message"`
	Data   json.RawMessage `json:"data"`
}

func pedir(t *testing.T, srv *httptest.Server, token, metodo, ruta string, body any) resp {
	t.Helper()
	var r io.Reader
	if body != nil {
		j, _ := json.Marshal(body)
		r = bytes.NewReader(j)
	}
	req, _ := http.NewRequest(metodo, srv.URL+ruta, r)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	out := resp{status: res.StatusCode}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatalf("%s %s: respuesta no JSON (%d): %v", metodo, ruta, res.StatusCode, err)
	}
	return out
}

// quiero falla si el status no es el esperado, y si hay destino le carga data.
func (r resp) quiero(t *testing.T, status int, destino any) {
	t.Helper()
	if r.status != status || r.Code != status {
		t.Fatalf("status %d (code %d), quería %d: %s", r.status, r.Code, status, r.Msg)
	}
	if destino != nil {
		if err := json.Unmarshal(r.Data, destino); err != nil {
			t.Fatal(err)
		}
	}
}

// entrar abre sesión con Google como hace el front; devuelve el token.
func entrar(t *testing.T, srv *httptest.Server, email string) string {
	t.Helper()
	var login struct {
		Token     string `json:"token"`
		TokenType string `json:"token_type"`
		User      struct {
			Email string `json:"email"`
		} `json:"user"`
	}
	pedir(t, srv, "", "POST", "/api/authentication/google", map[string]string{"credential": "google|sub-" + email + "|" + email}).quiero(t, 200, &login)
	if login.Token == "" || login.TokenType != "Bearer" || login.User.Email != email {
		t.Fatalf("login: %+v", login)
	}
	return login.Token
}

type notaT struct {
	ID   string `json:"_id"`
	Task string `json:"task"`
}
type columnaT struct {
	ID    string  `json:"_id"`
	Name  string  `json:"name"`
	Notes []notaT `json:"notes"`
}
type tableroT struct {
	ID      string     `json:"_id"`
	Name    string     `json:"name"`
	Rol     string     `json:"rol"`
	Columns []columnaT `json:"columns"`
}

func TestCuenta(t *testing.T) {
	srv := nuevaAPI(t)
	token := entrar(t, srv, "ana@mail.com")

	var yo struct {
		ID                    string `json:"_id"`
		Name, Email, Img, Rol string
	}
	pedir(t, srv, token, "GET", "/api/authentication/", nil).quiero(t, 200, &yo)
	if yo.ID != "sub-ana@mail.com" || yo.Name != "Ana" || yo.Email != "ana@mail.com" || yo.Img != "https://foto/ana.png" || yo.Rol != "user" {
		t.Fatalf("yo: %+v", yo)
	}

	// Volver a entrar no crea otro usuario: ve los mismos tableros.
	pedir(t, srv, token, "POST", "/api/work_space", map[string]string{"name": "T"}).quiero(t, 201, nil)
	otra := entrar(t, srv, "ana@mail.com")
	var lista []tableroT
	pedir(t, srv, otra, "GET", "/api/work_space", nil).quiero(t, 200, &lista)
	if len(lista) != 1 {
		t.Fatalf("al volver a entrar ve %d tableros", len(lista))
	}

	// Token de Google inválido, o el mismo email con otra cuenta de Google.
	pedir(t, srv, "", "POST", "/api/authentication/google", map[string]string{"credential": "inventado"}).quiero(t, 401, nil)
	pedir(t, srv, "", "POST", "/api/authentication/google", map[string]string{"credential": "google|otro-sub|ana@mail.com"}).quiero(t, 409, nil)

	// Salir cierra solo esa sesión.
	pedir(t, srv, token, "POST", "/api/authentication/logout", nil).quiero(t, 200, nil)
	pedir(t, srv, token, "GET", "/api/authentication", nil).quiero(t, 401, nil)
	pedir(t, srv, otra, "GET", "/api/authentication", nil).quiero(t, 200, nil)

	// Sin token, o con uno inventado, no entra.
	pedir(t, srv, "", "GET", "/api/work_space", nil).quiero(t, 401, nil)
	pedir(t, srv, "null", "GET", "/api/work_space", nil).quiero(t, 401, nil)
}

// El recorrido del front: crear tablero, columnas y notas, editarlas, moverlas y borrarlas.
func TestTablero(t *testing.T) {
	srv := nuevaAPI(t)
	token := entrar(t, srv, "ana@mail.com")
	p := func(metodo, ruta string, body any) resp { return pedir(t, srv, token, metodo, ruta, body) }
	tablero := func(id string) tableroT {
		var tb tableroT
		p("GET", "/api/work_space/"+id, nil).quiero(t, 200, &tb)
		return tb
	}

	var lista []tableroT
	p("GET", "/api/work_space", nil).quiero(t, 200, &lista)
	if lista == nil || len(lista) != 0 {
		t.Fatalf("sin tableros tiene que ser []: %+v", lista)
	}
	p("POST", "/api/work_space/", map[string]string{"user_id": "otro", "name": "Casa"}).quiero(t, 201, nil)
	p("POST", "/api/work_space/", map[string]string{"name": "  "}).quiero(t, 400, nil)
	p("GET", "/api/work_space", nil).quiero(t, 200, &lista)
	if len(lista) != 1 || lista[0].Name != "Casa" || lista[0].Rol != "owner" {
		t.Fatalf("lista: %+v", lista)
	}
	id := lista[0].ID
	if tb := tablero(id); tb.Columns == nil || len(tb.Columns) != 0 {
		t.Fatalf("un tablero nuevo tiene columns []: %+v", tb)
	}

	p("PATCH", "/api/work_space/", map[string]string{"_id": id, "name": "Mi casa"}).quiero(t, 200, nil)

	var a, b columnaT
	p("POST", "/api/column", map[string]string{"workspace_id": id, "name": "To do"}).quiero(t, 201, &a)
	p("POST", "/api/column", map[string]string{"workspace_id": id, "name": "Done"}).quiero(t, 201, &b)
	if a.Notes == nil {
		t.Fatal("una columna nueva tiene notes []")
	}
	// el front manda signal (un AbortSignal, llega como {}) y columnId de más
	p("PATCH", "/api/column/", map[string]any{"_id": b.ID, "name": "", "signal": map[string]any{}}).quiero(t, 200, nil)
	p("PATCH", "/api/column/", map[string]any{"_id": b.ID, "name": "Hecho", "signal": map[string]any{}}).quiero(t, 200, nil)

	var n1, n2, n3 notaT
	p("POST", "/api/note/", map[string]string{"column_id": a.ID, "task": "uno"}).quiero(t, 201, &n1)
	p("POST", "/api/note/", map[string]string{"column_id": a.ID, "task": "dos"}).quiero(t, 201, &n2)
	p("POST", "/api/note/", map[string]string{"column_id": a.ID, "task": "tres"}).quiero(t, 201, &n3)
	p("PATCH", "/api/note/", map[string]any{"_id": n1.ID, "task": "UNO", "columnId": a.ID, "signal": map[string]any{}}).quiero(t, 200, nil)

	// Mover "dos" a la otra columna: el front manda origen y destino; probamos los dos órdenes.
	p("PUT", "/api/column/", map[string]any{"_id": a.ID, "notes": []string{n3.ID, n1.ID}}).quiero(t, 200, nil)
	p("PUT", "/api/column/", map[string]any{"_id": b.ID, "notes": []string{n2.ID}}).quiero(t, 200, nil)
	p("PUT", "/api/column/", map[string]any{"_id": a.ID, "notes": []string{n2.ID, n3.ID, n1.ID}}).quiero(t, 200, nil)
	p("PUT", "/api/column/", map[string]any{"_id": b.ID, "notes": []string{}}).quiero(t, 200, nil)
	p("PUT", "/api/work_space/", map[string]any{"_id": id, "columns": []string{b.ID, a.ID}}).quiero(t, 200, nil)

	tb := tablero(id)
	if tb.Name != "Mi casa" || len(tb.Columns) != 2 || tb.Columns[0].Name != "Hecho" || len(tb.Columns[0].Notes) != 0 {
		t.Fatalf("tablero: %+v", tb)
	}
	if got := tb.Columns[1].Notes; len(got) != 3 || got[0].Task != "dos" || got[1].Task != "tres" || got[2].Task != "UNO" {
		t.Fatalf("notas: %+v", got)
	}

	p("DELETE", "/api/note/"+n3.ID, nil).quiero(t, 200, nil)
	p("DELETE", "/api/note/"+n3.ID, nil).quiero(t, 404, nil)
	p("DELETE", "/api/column/"+a.ID, nil).quiero(t, 200, nil)
	if tb := tablero(id); len(tb.Columns) != 1 || tb.Columns[0].ID != b.ID {
		t.Fatalf("después de borrar: %+v", tb)
	}

	p("DELETE", "/api/work_space/"+id, nil).quiero(t, 200, nil)
	p("GET", "/api/work_space/"+id, nil).quiero(t, 404, nil)
	p("GET", "/api/work_space", nil).quiero(t, 200, &lista)
	if len(lista) != 0 {
		t.Fatalf("quedó en la lista: %+v", lista)
	}
}

func TestValidaciones(t *testing.T) {
	srv := nuevaAPI(t)
	token := entrar(t, srv, "ana@mail.com")
	p := func(metodo, ruta string, body any) resp { return pedir(t, srv, token, metodo, ruta, body) }
	var tb tableroT
	var c columnaT
	p("POST", "/api/work_space", map[string]string{"name": "T"}).quiero(t, 201, &tb)
	p("POST", "/api/column", map[string]string{"workspace_id": tb.ID, "name": "C"}).quiero(t, 201, &c)

	p("POST", "/api/note", map[string]string{"column_id": c.ID, "task": string(make([]byte, 1001))}).quiero(t, 400, nil)
	p("PATCH", "/api/column", map[string]string{"_id": c.ID, "name": string(make([]byte, 101))}).quiero(t, 400, nil)
	p("POST", "/api/note", map[string]string{"column_id": tb.ID + "-NOEXISTE", "task": "x"}).quiero(t, 404, nil)
	p("PATCH", "/api/note", map[string]string{"_id": "cualquiera", "task": "x"}).quiero(t, 404, nil)
	p("POST", "/api/note", "no es un objeto").quiero(t, 400, nil)
}

// Otro usuario no ve ni toca tableros ajenos: para él no existen.
func TestTablerosAjenos(t *testing.T) {
	srv := nuevaAPI(t)
	ana, beto := entrar(t, srv, "ana@mail.com"), entrar(t, srv, "beto@mail.com")
	var tb tableroT
	var c columnaT
	var n notaT
	pedir(t, srv, ana, "POST", "/api/work_space", map[string]string{"name": "De Ana"}).quiero(t, 201, &tb)
	pedir(t, srv, ana, "POST", "/api/column", map[string]string{"workspace_id": tb.ID, "name": "C"}).quiero(t, 201, &c)
	pedir(t, srv, ana, "POST", "/api/note", map[string]string{"column_id": c.ID, "task": "x"}).quiero(t, 201, &n)

	for _, r := range []struct {
		metodo, ruta string
		body         any
	}{
		{"GET", "/api/work_space/" + tb.ID, nil},
		{"PATCH", "/api/work_space", map[string]string{"_id": tb.ID, "name": "mío"}},
		{"PUT", "/api/work_space", map[string]any{"_id": tb.ID, "columns": []string{}}},
		{"DELETE", "/api/work_space/" + tb.ID, nil},
		{"POST", "/api/column", map[string]string{"workspace_id": tb.ID, "name": "C"}},
		{"PATCH", "/api/column", map[string]string{"_id": c.ID, "name": "C"}},
		{"PUT", "/api/column", map[string]any{"_id": c.ID, "notes": []string{}}},
		{"DELETE", "/api/column/" + c.ID, nil},
		{"POST", "/api/note", map[string]string{"column_id": c.ID, "task": "x"}},
		{"PATCH", "/api/note", map[string]string{"_id": n.ID, "task": "x"}},
		{"DELETE", "/api/note/" + n.ID, nil},
	} {
		if res := pedir(t, srv, beto, r.metodo, r.ruta, r.body); res.status != 404 {
			t.Errorf("%s %s como beto: %d, quería 404", r.metodo, r.ruta, res.status)
		}
	}
	var lista []tableroT
	pedir(t, srv, beto, "GET", "/api/work_space", nil).quiero(t, 200, &lista)
	if len(lista) != 0 {
		t.Fatalf("beto ve: %+v", lista)
	}
	var tb2 tableroT
	pedir(t, srv, ana, "GET", "/api/work_space/"+tb.ID, nil).quiero(t, 200, &tb2)
	if len(tb2.Columns) != 1 || len(tb2.Columns[0].Notes) != 1 || tb2.Name != "De Ana" {
		t.Fatalf("beto cambió algo: %+v", tb2)
	}
}

// Muchos pedidos a la vez sobre el mismo tablero: ninguno se pierde.
func TestPedidosALaVez(t *testing.T) {
	srv := nuevaAPI(t)
	token := entrar(t, srv, "ana@mail.com")
	var tb tableroT
	var c columnaT
	pedir(t, srv, token, "POST", "/api/work_space", map[string]string{"name": "T"}).quiero(t, 201, &tb)
	pedir(t, srv, token, "POST", "/api/column", map[string]string{"workspace_id": tb.ID, "name": "C"}).quiero(t, 201, &c)

	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			if r := pedir(t, srv, token, "POST", "/api/note", map[string]string{"column_id": c.ID, "task": "x"}); r.status != 201 {
				t.Errorf("crear nota: %d %s", r.status, r.Msg)
			}
		})
	}
	wg.Wait()
	pedir(t, srv, token, "GET", "/api/work_space/"+tb.ID, nil).quiero(t, 200, &tb)
	if len(tb.Columns[0].Notes) != 8 {
		t.Fatalf("quedaron %d notas, quería 8", len(tb.Columns[0].Notes))
	}
}

func TestCorsYHealth(t *testing.T) {
	srv := nuevaAPI(t)
	req, _ := http.NewRequest("OPTIONS", srv.URL+"/api/work_space/", nil)
	req.Header.Set("Origin", "https://front.test")
	res, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 204 || res.Header.Get("Access-Control-Allow-Origin") != "https://front.test" {
		t.Fatalf("preflight: %d %v", res.StatusCode, res.Header)
	}
	res, err = srv.Client().Get(srv.URL + "/api/health")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != 200 || string(body) != "ok" {
		t.Fatalf("health: %d %s", res.StatusCode, body)
	}
}
