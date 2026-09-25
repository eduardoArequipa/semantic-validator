# Proyecto de código abierto

[English](OPEN_SOURCE.md) · Español

El [repositorio de GitHub](https://github.com/eduardoArequipa/semantic-validator)
contiene el servidor, los cuatro SDKs y la documentación. El código y los
paquetes descargables están bajo [Apache License 2.0](../LICENSE). Semantic
Validator es gratuito y de código abierto; no existe un plan alojado de pago.
Jev es un servicio aparte que cada desarrollador utiliza con su propia clave
y cuenta.

## Antes de publicar una versión o paquete

1. Confirmar la titularidad del código y revisar las licencias y avisos de las
   dependencias antes de publicar un repositorio o un paquete público.
2. El token Jev compartido anteriormente fue revocado y retirado de `.env`
   en el VPS. Revisar otras copias locales o historiales antes de publicar.
3. Ejecutar `go test ./...`, `go vet ./...`, pruebas de SDKs y
   `python3 tools/evaluate.py --dry-run`. El CI propuesto ejecuta estas pruebas
   al abrir el proyecto en GitHub; no necesita claves reales.
4. Revisar los archivos incluidos en Git. `.env`, datos de acceso, casos privados,
   binarios y artefactos generados están excluidos por `.gitignore`. La web
   sirve cuatro paquetes SDK de `internal/api/docs/downloads/`; revisar su
   contenido antes de publicar. El archivo de la antigua demo TypeScript
   permanece en el código, pero no se incorpora ni se sirve.
5. La ruta pública del módulo Go es
   `github.com/eduardoArequipa/semantic-validator`. Los SDKs TypeScript,
   Python y Java aún no están publicados en registros de paquetes. Considerar por
   separado las condiciones de uso y marca de Jev/TypeSafe al distribuir el
   proveedor y ofrecer el servicio alojado. El
   [acuerdo público de TypeSafe](https://typesafe.ai/legal/mca) prohíbe ofrecer
   sus servicios como un servicio independiente. Antes de ampliar el proxy
   alojado, solicitar permiso escrito. La ruta abierta recomendada es BYOK
   autohospedado, sujeta al contrato de cada operador.

## API alojada pausada

El VPS ejecuta `DOCS_ONLY`: la web y la documentación siguen disponibles,
pero `/v1/*` devuelve HTTP 503 sin consultar Jev. Los datos anteriores de
claves individuales permanecen guardados para recuperación; la línea del token
Jev revocado se retiró de `.env` en el VPS. No ampliar el acceso ni
promocionarla como proxy independiente sin permiso escrito de TypeSafe.
No hay registro autoservicio, facturación ni servicio alojado de pago.
