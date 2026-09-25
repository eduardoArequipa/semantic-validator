# Prueba los SDKs sin una clave de Jev

Clona el repositorio y entra en su raíz:

```bash
git clone https://github.com/eduardoArequipa/semantic-validator.git
cd semantic-validator
```

Cada carpeta contiene un programa que usa **tu propia clave** en un backend y
una prueba con respuestas simuladas. Empieza por la prueba: no contacta Jev
ni consume créditos. Después, si quieres probar el proveedor real, configura
`TYPESAFE_API_KEY` en el entorno de tu proceso; no la pongas en el código ni
en un navegador. Un archivo `.env` no se carga automáticamente salvo en el
demo TypeScript, que lo hace explícitamente.

| Lenguaje | Prueba sin clave | Ejecución real |
| --- | --- | --- |
| [TypeScript](typescript-demo/README.md) | `cd examples/typescript-demo && npm ci && npm test` | Desde esa carpeta: `npm run prueba` (lee `.env`) |
| [Python](python-direct/README.md) | `PYTHONPATH=sdk/python/src python3 -m unittest discover -s examples/python-direct -p 'test_*.py'` | `python3 examples/python-direct/main.py` (instala antes el SDK) |
| [Go](go-direct/README.md) | `go test ./examples/go-direct` | `go run ./examples/go-direct` |
| [Java](java-direct/README.md) | `mvn -f sdk/java/pom.xml install && mvn -f examples/java-direct/pom.xml test` | `mvn -f examples/java-direct/pom.xml compile exec:java` |

Para Python recomendamos el entorno virtual descrito en su README. En Java,
el primer comando instala el SDK en el Maven local antes de probar el ejemplo.
Los tests Go y Java abren **solo un servidor local** en la interfaz de loopback; ninguna
prueba necesita credenciales reales. El demo TypeScript ya incluye pruebas
simuladas para reclamos y formularios.

Cada ejecución real hace consultas a Jev y puede consumir créditos. `uncertain`
significa que no conviene decidir automáticamente; un error del proveedor no
significa `false`. La [guía de uso](https://validator.trialsur.cloud/docs/byok.html#uso)
explica cómo interpretar el resultado.
