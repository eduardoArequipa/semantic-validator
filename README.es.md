# Semantic Validator

[English](README.md) · Español · [Ejemplos de uso](https://validator.trialsur.cloud/docs/byok.html#uso) · [Reglas de campos](docs/RULES.es.md) · [Instalar una versión](docs/RELEASE.es.md) · [Resultados de evaluación](evals/RESULTS-2026-09-25.md)

Valida el significado de un texto con una API REST en Go y SDKs para Python,
TypeScript, Go y Java. Jev es el proveedor de inferencia semántica. Ejecutas
el servidor con **tu propia clave de Jev** (BYOK); el proyecto añade reglas,
resultados estructurados, caché, autenticación y cuotas.
Los cuatro SDKs ahora incluyen un cliente directo que llama a Jev con tu propia
clave sin ejecutar nuestro servidor. El cliente anterior para una instancia
propia sigue disponible.

La ruta `/` redirige a la documentación de SDKs en `/docs/`.
El código se distribuye bajo [Apache License 2.0](LICENSE) en el
[repositorio de GitHub](https://github.com/eduardoArequipa/semantic-validator).
Los SDKs TypeScript, Python y Java aún no están publicados en sus registros;
el SDK Go forma parte de este módulo público. Consulta
[las notas de código abierto](docs/OPEN_SOURCE.es.md) y la
[guía de SDKs directos](https://validator.trialsur.cloud/docs/byok.html) y la
[guía de servidor propio](docs/GETTING_STARTED.md).

## Uso directo sin nuestro servidor

Guarda `TYPESAFE_API_KEY` en un backend de confianza. En TypeScript:

```ts
import { DirectValidator } from "@semantic-validator/sdk";

const jevApiKey = process.env.TYPESAFE_API_KEY;
if (!jevApiKey) throw new Error("Falta TYPESAFE_API_KEY");
const validator = new DirectValidator({ jevApiKey });
const result = await validator.check("Mi pedido llegó dañado", "¿Es un reclamo?");
console.log(result.valid, result.confidence, result.status);
```

Python y Java exportan `DirectValidator`; Go ofrece `NewDirectClient`.
Cada consulta va directamente a Jev y se carga a tu cuenta. No hay clave,
cuota ni servidor de nuestra parte. No uses la clave en navegador o app móvil.

La licencia del código **no** incluye Jev ni autoriza revender su API. El
[acuerdo público de TypeSafe](https://typesafe.ai/legal/mca) limita la oferta de
sus servicios como servicio independiente. Antes de abrir un proxy alojado a
terceros, revisa tu contrato y solicita autorización escrita. Mientras tanto,
la instalación con tu propia clave es el camino recomendado.

Para evaluar calidad con ejemplos etiquetados y sin exponer textos en el
reporte, consulta [la guía de evaluación](evals/README.es.md).

## Configuración

```bash
cp .env.example .env
${EDITOR:-nano} .env
```

Completa `TYPESAFE_API_KEY` con tu token vigente de Jev. Luego carga las variables y ejecuta el servidor:

```bash
set -a
source .env
set +a
go run ./cmd/server
```

El token debe permanecer únicamente en `.env` o en un gestor de secretos. `.env` está excluido de Git.
`SEMANTIC_VALIDATOR_API_KEY` es una clave **distinta** para acceder a tu propia
instancia; cambia el valor `local-dev` antes de exponerla. Docker Compose local
solo publica en `127.0.0.1` por defecto.

## Ejecutar

```bash
go run ./cmd/server
```

## Validar un nombre

```bash
curl -X POST http://localhost:8080/v1/validate \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer local-dev' \
  -d '{"rule":"person_name","value":"Jorge Eduardo"}'
```

La respuesta contiene `valid`, `confidence` y `status`. Una confianza menor a `0.80` produce `status: "uncertain"` y `valid: null`.

Las validaciones repetidas durante la vida del proceso se sirven desde una caché segura en memoria. La caché se pierde al reiniciar el servidor.

## Salud y métricas

```bash
curl http://localhost:8080/health
curl http://localhost:8080/metrics
curl http://localhost:8080/openapi.yaml
```

El contrato completo de la API está en [`internal/api/openapi.yaml`](internal/api/openapi.yaml) y también se sirve desde `/openapi.yaml`.

Los benchmarks no realizan llamadas a Jev:

```bash
go test ./... -bench=. -benchmem
```

## Pregunta semántica personalizada

```bash
curl -X POST http://localhost:8080/v1/check \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer local-dev' \
  -d '{"value":"Necesito devolver el producto","question":"¿El cliente solicita una devolución?"}'
```

## Validación batch

```bash
curl -X POST http://localhost:8080/v1/validate/batch \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer local-dev' \
  -d '{"items":[{"id":"name","rule":"person_name","value":"Jorge Eduardo"},{"id":"other","rule":"person_name","value":"sdgfxcdg"}]}'
```

La respuesta conserva el orden de entrada. Cada elemento puede devolver un resultado o su propio error; el batch acepta entre 1 y 100 elementos.

## SDK Python

El SDK síncrono está en [`sdk/python`](sdk/python/README.md):

```bash
cd sdk/python
python -m pip install -e .
```

```python
from semantic_validator import Validator

validator = Validator(api_key="local-dev", base_url="http://localhost:8080")
result = validator.name("Jorge Eduardo")
print(result.valid, result.confidence)
```

## SDK TypeScript

El cliente TypeScript está en [`sdk/typescript`](sdk/typescript/README.md). Requiere Node.js 18 o superior:

```bash
cd sdk/typescript
npm install
npm run build
```

Usa `new Validator({ apiKey })` y ofrece `name`, `validate`, `check` y `validateBatch`.

## SDK Go

El cliente Go está en [`sdk/go`](sdk/go/README.md) y usa solo la biblioteca estándar. Al formar parte del módulo raíz, se importa como `github.com/eduardoArequipa/semantic-validator/sdk/go`.

## SDK Java

El cliente Java está en [`sdk/java`](sdk/java/README.md). Usa Java 11 o superior y Maven; incluye `name`, `validate`, `check` y `validateBatch`.

## Docker

Docker Compose carga las credenciales desde `.env`; ese archivo se excluye tanto de Git como del contexto de construcción de la imagen.

```bash
docker compose up --build -d
docker compose ps
curl http://localhost:8080/health
```

El contenedor ejecuta el servidor como usuario no privilegiado, con filesystem de solo lectura y health check. Para detenerlo:

```bash
docker compose down
```

La API exige `Authorization: Bearer <SEMANTIC_VALIDATOR_API_KEY>` en las rutas `/v1/*`. No uses aquí el token privado de Jev.

El servidor limita cada IP a 60 solicitudes por minuto en `/v1/*`. El límite se guarda en memoria y se reinicia al reiniciar el proceso. Por defecto identifica la IP de conexión y no confía en `X-Forwarded-For`.

## Despliegue en VPS con HTTPS

La configuración de producción de este repositorio mantiene la web pero
**pausa la API alojada** con `DOCS_ONLY=true`: `/v1/*` responde 503 y no se
necesita `TYPESAFE_API_KEY`. No la uses para ofrecer acceso público a Jev a
terceros sin revisar los términos del proveedor y obtener autorización cuando
corresponda. Para usar la API en tu aplicación, ejecuta tu propia instancia
con tu clave y con `DOCS_ONLY=false` (o sin definir esa variable).

La configuración de producción publica el API únicamente en `127.0.0.1:18081`, para que lo atienda el Nginx ya instalado en el VPS. Se usa el puerto 18081 porque el 8080 ya está ocupado por otro servicio. El archivo [`deploy/nginx/validator.trialsur.cloud.conf`](deploy/nginx/validator.trialsur.cloud.conf) es un nuevo sitio HTTP para `validator.trialsur.cloud`; no reemplaza ni modifica otros sitios. No cambies los registros de `trialsur.cloud` ni de `api.trialsur.cloud`.

El Compose de producción habilita `TRUST_PROXY_HEADERS=true`: solo es seguro porque el servicio está publicado en `127.0.0.1`, detrás del Nginx del mismo VPS. El limitador valida la última IP de `X-Forwarded-For`, que Nginx añade a la cadena; en desarrollo, la opción permanece apagada por defecto.

El DNS del subdominio debe apuntar al VPS. En este caso Nginx ya ocupa 80/443, por lo que no levantes otro proxy que intente reservar esos puertos.

En el VPS, copia el proyecto y crea `.env` para los valores de puerto que uses;
en modo solo documentación no hace falta ninguna clave. No compartas ni subas
ese archivo. Arranca el sitio:

```bash
docker compose -f docker-compose.prod.yml up --build -d
docker compose -f docker-compose.prod.yml ps
curl http://127.0.0.1:18081/health
```

Después instala el sitio Nginx sin sobrescribir archivos existentes y valida la configuración antes de recargar Nginx:

```bash
install -m 0644 deploy/nginx/validator.trialsur.cloud.conf /etc/nginx/sites-available/validator.trialsur.cloud
ln -s /etc/nginx/sites-available/validator.trialsur.cloud /etc/nginx/sites-enabled/validator.trialsur.cloud
nginx -t && systemctl reload nginx
```

Cuando HTTP responda correctamente, habilita HTTPS con el mecanismo Certbot ya instalado en el servidor (por ejemplo, `certbot --nginx -d validator.trialsur.cloud`) y comprueba `curl https://validator.trialsur.cloud/health`. Para los logs del API: `docker compose -f docker-compose.prod.yml logs -f semantic-validator`.


## Claves individuales para testers

El siguiente apartado documenta la beta anterior. En el VPS actual,
`DOCS_ONLY=true` desactiva todas las validaciones, aunque los datos de claves
permanezcan guardados para recuperación. Este modo técnico solo es para
instalaciones autorizadas; no implica permiso para redistribuir Jev.

La web pública ahora contiene solo la
[documentación de SDKs](https://validator.trialsur.cloud/docs/byok.html):
instalación de una instancia propia, API REST y descargas de los cuatro SDKs.
La portada anterior y la demo alojada dejaron de servirse. Los archivos
necesarios se incorporan al binario mediante `go:embed`.

Para probar compra y reclamos con el SDK, consulta la
[demo TypeScript para testers](examples/typescript-demo/README.md). Incluye carga
de `.env`, casos de referencia y un formulario de feedback sin datos personales.

La API admite claves revocables con cuota diaria persistente por clave. El comando
`cmd/keys` permite crearlas, consultar consumo y revocarlas. Un batch reserva una
unidad por elemento antes de comenzar; las llamadas intentadas a Jev se contabilizan
por separado de las validaciones servidas desde caché.

Consulta [activación y administración](deploy/individual-keys.md). El modo se activa
con `API_KEYS_FILE`; al activarlo deja de aceptarse la clave global anterior. Los
SDKs mantienen su interfaz: cada tester configura su propia `apiKey`.

## Pruebas

```bash
go test ./...
go vet ./...
```
