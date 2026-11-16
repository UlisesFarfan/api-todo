package base

import (
	"context"
	"errors"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Dynamo es la base en AWS: una tabla con clave pk + sk (ver template.yaml).
// Con AWS_ENDPOINT_URL_DYNAMODB apunta a DynamoDB Local.
type Dynamo struct {
	c     *dynamodb.Client
	tabla string
}

func NuevaDynamo(ctx context.Context, tabla string) (*Dynamo, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &Dynamo{c: dynamodb.NewFromConfig(cfg), tabla: tabla}, nil
}

func s(v string) types.AttributeValue { return &types.AttributeValueMemberS{Value: v} }
func n(v int64) types.AttributeValue {
	return &types.AttributeValueMemberN{Value: strconv.FormatInt(v, 10)}
}

func clave(pk, sk string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{"pk": s(pk), "sk": s(sk)}
}

func aFila(item map[string]types.AttributeValue) Fila {
	var f Fila
	if v, ok := item["pk"].(*types.AttributeValueMemberS); ok {
		f.PK = v.Value
	}
	if v, ok := item["sk"].(*types.AttributeValueMemberS); ok {
		f.SK = v.Value
	}
	if v, ok := item["datos"].(*types.AttributeValueMemberB); ok {
		f.Datos = v.Value
	}
	if v, ok := item["version"].(*types.AttributeValueMemberN); ok {
		f.Version, _ = strconv.ParseInt(v.Value, 10, 64)
	}
	if v, ok := item["vence"].(*types.AttributeValueMemberN); ok {
		f.Vence, _ = strconv.ParseInt(v.Value, 10, 64)
	}
	return f
}

func (d *Dynamo) Leer(ctx context.Context, pk, sk string) (Fila, bool, error) {
	out, err := d.c.GetItem(ctx, &dynamodb.GetItemInput{TableName: &d.tabla, Key: clave(pk, sk), ConsistentRead: aws.Bool(true)})
	if err != nil || out.Item == nil {
		return Fila{}, false, err
	}
	return aFila(out.Item), true, nil
}

func (d *Dynamo) Consultar(ctx context.Context, pk, prefijo string) ([]Fila, error) {
	p := dynamodb.NewQueryPaginator(d.c, &dynamodb.QueryInput{
		TableName:                 &d.tabla,
		KeyConditionExpression:    aws.String("pk = :pk AND begins_with(sk, :p)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{":pk": s(pk), ":p": s(prefijo)},
		ConsistentRead:            aws.Bool(true),
	})
	var filas []Fila
	for p.HasMorePages() {
		out, err := p.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, item := range out.Items {
			filas = append(filas, aFila(item))
		}
	}
	return filas, nil
}

func (d *Dynamo) Escribir(ctx context.Context, cambios ...Cambio) error {
	items := make([]types.TransactWriteItem, 0, len(cambios))
	for _, c := range cambios {
		var cond *string
		valores := map[string]types.AttributeValue{}
		switch {
		case c.Version != 0:
			cond = aws.String("version = :v")
			valores[":v"] = n(c.Version)
		case !c.Borrar:
			cond = aws.String("attribute_not_exists(pk)")
		}
		if len(valores) == 0 {
			valores = nil
		}
		if c.Borrar {
			items = append(items, types.TransactWriteItem{Delete: &types.Delete{
				TableName: &d.tabla, Key: clave(c.PK, c.SK), ConditionExpression: cond, ExpressionAttributeValues: valores}})
			continue
		}
		item := clave(c.PK, c.SK)
		item["datos"] = &types.AttributeValueMemberB{Value: c.Datos}
		item["version"] = n(c.Version + 1)
		if c.Vence != 0 {
			item["vence"] = n(c.Vence)
		}
		items = append(items, types.TransactWriteItem{Put: &types.Put{
			TableName: &d.tabla, Item: item, ConditionExpression: cond, ExpressionAttributeValues: valores}})
	}
	_, err := d.c.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: items})
	var cancelada *types.TransactionCanceledException
	if errors.As(err, &cancelada) {
		for _, r := range cancelada.CancellationReasons {
			// TransactionConflict: otra transacción tocaba la misma fila; se reintenta igual.
			if c := aws.ToString(r.Code); c == "ConditionalCheckFailed" || c == "TransactionConflict" {
				return ErrConflicto
			}
		}
	}
	return err
}
