# Desarrollo local (la primera vez: cp .env.local.example .env.local):
#   make api      la API en http://localhost:8080 (sin TABLA, los datos quedan en memoria)
#   make test     vet + tests
#
# Publicar en AWS (cp .env.prod.example .env.prod y completar; ver README, "Publicar en AWS"):
#   make deploy   compila, despliega con SAM y muestra la URL para VITE_API_ENDPOINT

-include .env.local
export

.PHONY: api test build deploy

api:
	go run ./cmd/api

test:
	go vet ./...
	go test ./...

build:
	mkdir -p build
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -o build/bootstrap ./cmd/lambda

# Cada receta carga .env.prod en su propio shell.
PROD = set -a && . ./.env.prod && set +a

deploy: build
	$(PROD) && sam deploy --template-file template.yaml --stack-name "$$STACK_NAME" --region "$$AWS_DEFAULT_REGION" \
	  --capabilities CAPABILITY_IAM --resolve-s3 --no-confirm-changeset --no-fail-on-empty-changeset \
	  --parameter-overrides AllowedOrigins="$$ALLOWED_ORIGINS" GoogleClientId="$$GOOGLE_CLIENT_ID"
	$(PROD) && echo "VITE_API_ENDPOINT=$$(aws cloudformation describe-stacks --stack-name "$$STACK_NAME" \
	  --query "Stacks[0].Outputs[?OutputKey=='ApiUrl'].OutputValue" --output text)"
