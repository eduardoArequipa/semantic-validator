# Semantic Validator: SDK Python

[English](README.md) · [Ejemplo ejecutable](https://github.com/eduardoArequipa/semantic-validator/tree/main/examples/python-direct)

SDK síncrono para validar el significado de campos y hacer preguntas semánticas
con Jev. Requiere Python 3.10 o superior, no tiene dependencias de ejecución
y está publicado bajo Apache-2.0. Cada desarrollador utiliza su propia cuenta
y clave de Jev.

## Instalación

El nombre del paquete es `semantic-validator`; en Python se importa como
`semantic_validator`. El comando de instalación desde el registro es:

```bash
python -m pip install semantic-validator
```

Para instalarlo desde el repositorio:

```bash
python3 -m venv .venv
source .venv/bin/activate
python -m pip install -e sdk/python
```

## Uso directo con Jev

Configura `TYPESAFE_API_KEY` en el entorno de tu backend. Los archivos `.env`
no se cargan automáticamente.

```python
import os
from semantic_validator import DirectValidator, SemanticValidatorError

validator = DirectValidator(jev_api_key=os.environ["TYPESAFE_API_KEY"])

try:
    result = validator.check("Mi pedido llegó roto", "¿Es un reclamo?")
    if result.status == "uncertain":
        print("Requiere revisión", result.confidence)
    else:
        print(result.valid, result.confidence, result.status)
except SemanticValidatorError as error:
    print("Falló la consulta al proveedor:", error.code)
```

También ofrece `name`, `validate` y `validate_batch`. Las reglas de campos
son `person_name`, `product_description` y `address` (v1); consulta sus
[definiciones y límites](https://github.com/eduardoArequipa/semantic-validator/blob/main/docs/RULES.es.md).

Cada consulta envía el texto directamente a Jev y puede consumir créditos.
Con confianza inferior a 0.80, `valid` es `None` y `status` es `uncertain`.
Un error del proveedor no es un resultado inválido. Mantén tu clave en un
backend de confianza.

## Servidor propio

```python
import os
from semantic_validator import Validator

validator = Validator(
    api_key=os.environ["SEMANTIC_VALIDATOR_API_KEY"],
    base_url="http://localhost:8080",
)
result = validator.name("Jorge Eduardo")
print(result.valid, result.confidence, result.status)
```

`api_key` debe coincidir con la clave de tu instancia de Semantic Validator.
Consulta la [guía de servidor propio](https://github.com/eduardoArequipa/semantic-validator/blob/main/docs/GETTING_STARTED.md).
