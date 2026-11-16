package base

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func TestMemoria(t *testing.T) { probarBase(t, NuevaMemoria()) }

// TestDynamo corre lo mismo contra DynamoDB Local (ver README, "Tests"); sin
// AWS_ENDPOINT_URL_DYNAMODB se saltea.
func TestDynamo(t *testing.T) {
	if os.Getenv("AWS_ENDPOINT_URL_DYNAMODB") == "" {
		t.Skip("sin AWS_ENDPOINT_URL_DYNAMODB")
	}
	ctx := context.Background()
	tabla := "test-" + time.Now().Format("150405.000")
	d, err := NuevaDynamo(ctx, tabla)
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.c.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName:   &tabla,
		BillingMode: types.BillingModePayPerRequest,
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("pk"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("sk"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("pk"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("sk"), KeyType: types.KeyTypeRange},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.c.DeleteTable(ctx, &dynamodb.DeleteTableInput{TableName: &tabla}) })
	probarBase(t, d)
}

func probarBase(t *testing.T, b Base) {
	ctx := context.Background()
	nueva := Cambio{Fila: Fila{PK: "A", SK: "T#2", Datos: []byte(`{"x":1}`), Vence: 99}}

	if err := b.Escribir(ctx, nueva); err != nil {
		t.Fatal(err)
	}
	if err := b.Escribir(ctx, nueva); !errors.Is(err, ErrConflicto) {
		t.Fatalf("crear dos veces: %v, quería ErrConflicto", err)
	}
	f, ok, err := b.Leer(ctx, "A", "T#2")
	if err != nil || !ok || f.Version != 1 || string(f.Datos) != `{"x":1}` || f.Vence != 99 {
		t.Fatalf("leer: %+v %v %v", f, ok, err)
	}

	// Con la versión vieja no se guarda; con la que se leyó, sí.
	f.Datos = []byte(`{"x":2}`)
	if err := b.Escribir(ctx, f.conVersion(5)); !errors.Is(err, ErrConflicto) {
		t.Fatalf("versión vieja: %v", err)
	}
	if err := b.Escribir(ctx, Cambio{Fila: f}); err != nil {
		t.Fatal(err)
	}

	// Todo o nada: si una falla, la otra no se escribe.
	otra := Cambio{Fila: Fila{PK: "A", SK: "T#1", Datos: []byte(`{}`)}}
	if err := b.Escribir(ctx, otra, nueva); !errors.Is(err, ErrConflicto) {
		t.Fatalf("transacción: %v", err)
	}
	if _, ok, _ := b.Leer(ctx, "A", "T#1"); ok {
		t.Fatal("la transacción fallida escribió T#1")
	}
	if err := b.Escribir(ctx, otra, Cambio{Fila: Fila{PK: "A", SK: "X", Datos: []byte(`{}`)}}); err != nil {
		t.Fatal(err)
	}

	filas, err := b.Consultar(ctx, "A", "T#")
	if err != nil || len(filas) != 2 || filas[0].SK != "T#1" || filas[1].SK != "T#2" || string(filas[1].Datos) != `{"x":2}` {
		t.Fatalf("consultar: %+v %v", filas, err)
	}

	// Borrar con versión vieja falla; sin versión borra siempre.
	if err := b.Escribir(ctx, Cambio{Fila: Fila{PK: "A", SK: "T#2", Version: 1}, Borrar: true}); !errors.Is(err, ErrConflicto) {
		t.Fatalf("borrar con versión vieja: %v", err)
	}
	if err := b.Escribir(ctx, Cambio{Fila: Fila{PK: "A", SK: "T#2"}, Borrar: true}); err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := b.Leer(ctx, "A", "T#2"); ok {
		t.Fatal("no se borró")
	}
}

func (f Fila) conVersion(v int64) Cambio { f.Version = v; return Cambio{Fila: f} }
