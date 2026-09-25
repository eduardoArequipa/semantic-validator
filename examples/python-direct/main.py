"""A small BYOK example. Run only with your own key in a trusted backend."""

import os
import sys

from semantic_validator import DirectValidator, SemanticValidatorError


def show(label, result, output=print):
    if result.status == "uncertain":
        output(f"{label}: requiere revisión (confianza {result.confidence:.2f})")
    else:
        output(f"{label}: {result.valid} (confianza {result.confidence:.2f})")


def run(validator, output=print):
    show("nombre", validator.name("Jorge Eduardo"), output)
    show(
        "reclamo",
        validator.check(
            "Mi pedido llegó dañado y quiero una solución",
            "¿El cliente está presentando un reclamo?",
        ),
        output,
    )


def main():
    key = os.environ.get("TYPESAFE_API_KEY", "").strip()
    if not key:
        print("Falta TYPESAFE_API_KEY; ejecuta primero la prueba sin clave.", file=sys.stderr)
        return 2
    try:
        run(DirectValidator(jev_api_key=key))
    except SemanticValidatorError as error:
        print(f"Jev no pudo responder ({error.code}); no es un resultado inválido.", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
