# Evaluar Semantic Validator

[English](README.md) · Español

Para comparar Jev con Laya sin cambiar producción, consulta la
[guía del experimento de 100 casos](PROVIDER_COMPARISON.md). Puedes validar
el conjunto sin llamadas de red con
`python3 tools/compare_providers.py --dry-run`; para la ejecución real necesitas
Laya corriendo localmente y una clave de Semantic Validator con al menos 100
validaciones disponibles.

Esta evaluación compara respuestas de **Jev directo o tu propia API** con ejemplos
etiquetados antes de la ejecución. Los diez ejemplos originales son ficticios
e ilustrativos. [`field-rules.json`](field-rules.json) añade 30 casos ficticios
por regla (12 positivos, 12 negativos y 6 ambiguos). Son ejemplos iniciales
de regresión, no una medición de precisión. Cada solicitud real puede consumir
créditos de Jev y, en modo de claves individuales, una unidad de cuota.
`--dry-run` no realiza solicitudes.

```bash
python3 tools/evaluate.py --dry-run
python3 tools/evaluate.py --cases evals/field-rules.json --dry-run
```

La [prueba piloto de nueve casos](PILOT-2026-09-25.md) está documentada por
separado; no representa una medición de calidad en producción.
La [ejecución completa de 90 casos](RESULTS-2026-09-25.md) publica resultados
agregados, método y limitaciones sin publicar claves ni textos privados.

Para una prueba real, copia `evals/cases.example.json` a
`evals/private/cases.json` y añade ejemplos ficticios o autorizados de tu caso
de uso. El directorio privado está excluido de Git. Cada entrada contiene un
`id` no sensible, `value`, `expected_valid` (`true`, `false` o `null` para
casos ambiguos) y exactamente uno
de `rule` o `question`. Etiqueta el resultado esperado antes de consultar la API.
Usa la misma pregunta para los ejemplos que quieras comparar.

Para una evaluación real recomendamos llamar directamente a Jev con tu propia
cuenta. Instala el SDK Python en un entorno virtual e introduce tu clave sin
guardarla en el historial de comandos. Una ejecución completa de 90 casos
puede consumir **90 consultas a Jev**; empieza con `--sample-per-label 1`,
una muestra equilibrada de nueve consultas (un caso positivo, negativo y
ambiguo por regla):

```bash
python -m pip install -e sdk/python
read -rsp 'Clave Jev: ' TYPESAFE_API_KEY
printf '\n'
export TYPESAFE_API_KEY
python3 tools/evaluate.py --direct --cases evals/field-rules.json --sample-per-label 1
unset TYPESAFE_API_KEY
```

La API alojada está pausada. Si prefieres evaluar mediante tu propio servidor
Go, inícialo con tu clave Jev según [la guía](../docs/GETTING_STARTED.md). En
otra terminal usa la clave local de `.env.example` o la de tu instancia:

```bash
SEMANTIC_VALIDATOR_API_KEY=local-dev python3 tools/evaluate.py \
  --cases evals/field-rules.json --base-url http://localhost:8080
```

El informe muestra ID, respuesta, confianza, falsos positivos, falsos negativos,
inciertos, casos ambiguos y errores, en total y por regla. No imprime textos ni claves y no guarda respuestas. Si
encuentras un falso positivo o negativo, revisa el caso y la pregunta; un fallo
HTTP no se cuenta como decisión semántica. Repite la evaluación tras cambiar una
regla o el proveedor, usando el mismo conjunto etiquetado. No publiques ejemplos
privados ni los envíes a esta API sin autorización para procesarlos con Jev.

Históricamente, en una prueba de los diez ejemplos originales contra la API
alojada (antes de pausarla) el
23 de septiembre de 2026, nueve coincidieron con su etiqueta, uno quedó
incierto y no hubo errores HTTP. El caso incierto era una solicitud de soporte
vaga en inglés. Esta muestra no demuestra igual calidad en distintos idiomas.

Prueba local del evaluador:

```bash
python3 -m unittest discover -s tools -p 'test_*.py'
```
