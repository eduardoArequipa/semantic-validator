# Reglas de campos (beta)

[English](RULES.md) · [Guía de evaluación](../evals/README.es.md)

La fuente versionada está en [`internal/rules/catalog.json`](../internal/rules/catalog.json).
Tu servidor `/v1/validate` y `validate(rule, value)` de los SDKs directos usan
las mismas preguntas. Si se cambia una pregunta, debe cambiar la versión de la
regla para no reutilizar resultados de caché o evaluaciones anteriores. El
servidor incluye esa versión en la clave de caché; los SDKs directos no usan
nuestra caché.

| Regla (v1) | Sí previsto | No previsto | Revisar |
| --- | --- | --- | --- |
| `person_name` | Parece un nombre personal | Texto aleatorio, empresa, producto o cargo | Una sola palabra que también puede ser común |
| `product_description` | Producto y al menos una característica concreta | Elogio genérico o eslogan | Producto con característica vaga o ausente |
| `address` | Lugar físico suficientemente específico | Solo país o ciudad, texto no relacionado | Referencias incompletas o solo un punto de referencia |

Estas reglas juzgan si un texto *parece* cumplir la condición. No comprueban
la identidad de una persona, la veracidad de las características de un producto
ni si una dirección existe o admite entregas. Una entrada vacía o demasiado
larga es un error de solicitud; `invalid` es una decisión semántica,
`uncertain` pide revisión y un fallo del proveedor es un error. La confianza
de Jev **no** es una probabilidad calibrada de acierto. Decide cómo tratar
inciertos y errores antes de bloquear usuarios con estas reglas.

El conjunto público [`field-rules.json`](../evals/field-rules.json) contiene
30 casos ficticios por regla: 12 positivos, 12 negativos y 6 ambiguos. `null`
marca los ambiguos para revisión, no como aciertos o errores. Son ejemplos
iniciales de regresión, no una medición de precisión en producción. Ejecuta
una evaluación real con tu clave Jev y revisa los errores antes de confiar en
una regla.
