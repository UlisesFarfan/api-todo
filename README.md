# API de todo-app

La API del tablero de tareas (workspaces, columnas y notas). El front (`app-todo`, React)
está en Vercel y le habla con JSON `{code, status, message, data}` (ids en `_id`) y
`Authorization: Bearer <token>`.

Se entra solo con Google: el front manda el ID token del botón de Google a
`POST /api/authentication/google`, la API verifica firma, audiencia (`GOOGLE_CLIENT_ID`),
emisor, vencimiento y email verificado, crea el usuario la primera vez y abre una sesión.
`POST /api/authentication/logout` la cierra.

Armada como los productos gIA: Go solo con la biblioteca estándar (`net/http`), handlers con
validación y mensajes para el usuario, sesiones guardadas en la base, y despliegue en AWS con
SAM (Lambda arm64 + API Gateway). La base es **DynamoDB**: una sola tabla (ver `internal/base`).

- `cmd/api`: servidor local · `cmd/lambda`: el mismo router en Lambda (runtime propio, sin SDK).
- `internal/router`: todas las rutas en un lugar, y los tests que usan la API entera por HTTP.
- `internal/handlers`: auth (Google, yo, salir) y tableros (workspaces, columnas, notas).
- `internal/base`: la tabla (DynamoDB en AWS, memoria en local y en los tests).

## Datos

| Fila (pk / sk)               | Qué guarda                                                  |
|------------------------------|-------------------------------------------------------------|
| `U#<email>` / `PERFIL`       | el usuario (id = `sub` de su cuenta de Google)              |
| `U#<email>` / `T#<id>`       | que ese usuario es miembro del tablero (para listar)        |
| `T#<id>` / `TABLERO`         | el tablero entero: nombre, miembros, columnas y notas       |
| `S#<sha256 del token>` / `SESION` | una sesión de 7 días; DynamoDB la borra al vencer (TTL) |

Cada tablero es una fila porque el front siempre lo pide entero. Las columnas y notas llevan el
id del tablero adelante (`<tablero>-<id>`), así se encuentran sin índices. Toda escritura
lleva la versión de la fila: si dos pedidos tocan el mismo tablero a la vez (el front manda
dos al mover una nota de columna), el segundo vuelve a leer y reintenta.

Un usuario solo ve y toca sus tableros; para los ajenos la API responde 404.

## Levantarlo en local

```sh
cp .env.local.example .env.local   # la primera vez
make api                           # http://localhost:8080, datos en memoria
```

`curl localhost:8080/api/health` responde `ok`. En el front local, `VITE_API_ENDPOINT=http://localhost:8080`.

## Publicar en AWS

Requisitos: AWS CLI y SAM CLI configurados, y Go.

```sh
cp .env.prod.example .env.prod     # la primera vez: perfil, región, stack, origen del front, client ID de Google
make deploy                        # compila, despliega (tabla + Lambda + API Gateway) y muestra la URL
```

Después, en Vercel: `VITE_API_ENDPOINT=<la URL que mostró>` (sin barra final) y volver a
desplegar el front. Si cambia el dominio del front, actualizar `ALLOWED_ORIGINS` y `make deploy`.

La tabla tiene `DeletionPolicy: Retain`: borrar el stack no borra los datos.

## Tests

```sh
make test
```

Los de `internal/router` levantan la API entera contra la base en memoria y la usan como el
front (con un Google falso): login, salir, el recorrido de un tablero, mover notas entre columnas en los dos
órdenes, tableros ajenos (404) y pedidos a la vez.

`internal/base` prueba lo mismo sobre memoria y sobre DynamoDB. El de DynamoDB se saltea si no
hay `AWS_ENDPOINT_URL_DYNAMODB`; para correrlo, con
[DynamoDB Local](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DynamoDBLocal.DownloadingAndRunning.html)
en el puerto 8000:

```sh
AWS_ENDPOINT_URL_DYNAMODB=http://localhost:8000 AWS_ACCESS_KEY_ID=local AWS_SECRET_ACCESS_KEY=local \
AWS_REGION=sa-east-1 go test -run TestDynamo -v ./internal/base
```
