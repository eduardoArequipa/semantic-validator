# Semantic Validator Python SDK

SDK síncrono sin dependencias externas: modo directo a Jev y cliente para API propia.

## Modo directo (recomendado)

```python
import os
from semantic_validator import DirectValidator

validator = DirectValidator(jev_api_key=os.environ["TYPESAFE_API_KEY"])
result = validator.check("Mi pedido llegó roto", "¿Es un reclamo?")
print(result.valid, result.confidence, result.status)
```

Úsalo solo en un backend. También ofrece `name`, `validate` y `validate_batch`; cada elemento consulta Jev por separado. Con confianza inferior a 0.80, `valid` es `None`.

## Instalación local

```bash
cd sdk/python
python -m pip install -e .
```

## Servidor propio (opcional)

```python
from semantic_validator import Validator

validator = Validator(
    api_key="local-dev",
    base_url="http://localhost:8080",
)

result = validator.name("Jorge Eduardo")
print(result.valid)
print(result.confidence)

result = validator.check(
    "Necesito devolver el producto porque llegó roto",
    "¿El cliente está solicitando una devolución?",
)
print(result.status)
```

`api_key` es la clave de Semantic Validator. En desarrollo debe coincidir con `SEMANTIC_VALIDATOR_API_KEY` del servidor. No uses aquí el token privado de Jev.

## Batch

```python
results = validator.validate_batch([
    {"id": "name", "rule": "person_name", "value": "Jorge Eduardo"},
    {"id": "other", "rule": "person_name", "value": "sdgfxcdg"},
])
```

El SDK convierte errores HTTP y errores de transporte en `SemanticValidatorError`.

## Tests

```bash
python -m unittest discover -s tests
```
