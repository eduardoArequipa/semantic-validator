# Evaluar Semantic Validator

[English](README.md) · Español

Para comparar Jev con Laya sin cambiar producción, consulta la
[guía del experimento de 100 casos](PROVIDER_COMPARISON.md). Puedes validar
el conjunto sin llamadas de red con
`python3 tools/compare_providers.py --dry-run`; para la ejecución real necesitas
Laya corriendo localmente y una clave de Semantic Validator con al menos 100
validaciones disponibles.

Esta evaluación compara respuestas de la API con ejemplos etiquetados por una
persona. La muestra incluye español e inglés; es ficticia e ilustrativa. Diez
casos no sirven para estimar calidad general. Cada solicitud ejecutada puede consumir una unidad de
cuota, incluso si se sirve desde caché. `--dry-run` no realiza solicitudes.

```bash
python3 tools/evaluate.py --dry-run
```

Para una prueba real, copia `evals/cases.example.json` a
`evals/private/cases.json` y añade ejemplos ficticios o autorizados de tu caso
de uso. El directorio privado está excluido de Git. Cada entrada contiene un
`id` no sensible, `value`, `expected_valid` (`true` o `false`) y exactamente uno
de `rule` o `question`. Etiqueta el resultado esperado antes de consultar la API.
Usa la misma pregunta para los ejemplos que quieras comparar.

Desde la terminal, carga una clave individual sin escribirla en el historial:

```bash
read -rsp 'API key: ' SEMANTIC_VALIDATOR_API_KEY
printf '\n'
export SEMANTIC_VALIDATOR_API_KEY
python3 tools/evaluate.py --cases evals/private/cases.json \
  --base-url https://validator.trialsur.cloud
unset SEMANTIC_VALIDATOR_API_KEY
```

El informe muestra ID, respuesta, confianza, falsos positivos, falsos negativos,
inciertos y errores. No imprime textos ni claves y no guarda respuestas. Si
encuentras un falso positivo o negativo, revisa el caso y la pregunta; un fallo
HTTP no se cuenta como decisión semántica. Repite la evaluación tras cambiar una
regla o el proveedor, usando el mismo conjunto etiquetado. No publiques ejemplos
privados ni los envíes a esta API sin autorización para procesarlos con Jev.

En una prueba de estos diez ejemplos ficticios contra la API alojada el
23 de septiembre de 2026, nueve coincidieron con su etiqueta, uno quedó
incierto y no hubo errores HTTP. El caso incierto era una solicitud de soporte
vaga en inglés. Esta muestra no demuestra igual calidad en distintos idiomas.

Prueba local del evaluador:

```bash
python3 -m unittest discover -s tools -p 'test_*.py'
```
