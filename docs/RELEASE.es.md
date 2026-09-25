# Instalar una versión etiquetada

Empieza en la [versión v0.3.0](https://github.com/eduardoArequipa/semantic-validator/releases/tag/v0.3.0).
Descarga el SDK que necesitas y `SHA256SUMS.txt`. Si descargas los cuatro
paquetes, verifica todos con `sha256sum -c SHA256SUMS.txt`; si descargas solo
uno, compara su SHA-256 con la línea correspondiente. Nunca incluyas tu clave
de Jev en el frontend ni en Git.

## TypeScript / Node.js

Requiere Node.js 18+ en un backend de confianza. Descarga
`semantic-validator-sdk-0.3.0.tgz` y ejecuta:

```bash
npm install ./semantic-validator-sdk-0.3.0.tgz
```

Importa `DirectValidator` desde `@semantic-validator/sdk`. Consulta la
[guía de TypeScript](../sdk/typescript/README.md). Aún no está en npm.

## Python

Requiere Python 3.10+ y un entorno virtual. Descarga
`semantic-validator-python-0.3.0.tar.gz` y ejecuta:

```bash
tar -xzf semantic-validator-python-0.3.0.tar.gz
python3 -m venv .venv
. .venv/bin/activate
python -m pip install ./semantic-validator/sdk/python
```

Importa `DirectValidator` desde `semantic_validator`. Consulta la
[guía de Python](../sdk/python/README.md). Aún no está en PyPI.

## Go

El SDK Go forma parte del módulo etiquetado:

```bash
go get github.com/eduardoArequipa/semantic-validator/sdk/go@v0.3.0
```

Importa `github.com/eduardoArequipa/semantic-validator/sdk/go` y utiliza
`validator.NewDirectClient`. Consulta la [guía de Go](../sdk/go/README.md).
La versión también incluye un archivo de código fuente del SDK.

## Java

Requiere Java 11+ y Maven. Descarga
`semantic-validator-java-0.3.0.tar.gz` y ejecuta:

```bash
tar -xzf semantic-validator-java-0.3.0.tar.gz
mvn -f semantic-validator/sdk/java/pom.xml install
```

Luego usa `io.semanticvalidator:semantic-validator-sdk:0.3.0` en tu proyecto.
Esto lo instala en tu Maven **local**; aún no está en Maven Central. Consulta
la [guía de Java](../sdk/java/README.md).

## Antes de usarlo en producción

Guarda `TYPESAFE_API_KEY` solo en un backend de confianza. Los SDKs directos
consultan Jev con tu cuenta; nuestra API alojada sigue pausada. Un resultado
`uncertain` no significa `false`, y un error del proveedor tampoco.
La [evaluación publicada](../evals/RESULTS-2026-09-25.md) usa pocos casos
ficticios y no demuestra precisión en datos reales.
