# Contribuir

[English](CONTRIBUTING.md) · Español

Gracias por querer mejorar Semantic Validator. El código se distribuye bajo
[Apache License 2.0](LICENSE). Puedes contribuir en el
[repositorio de GitHub](https://github.com/eduardoArequipa/semantic-validator).

## Preparación

- Go 1.22 o posterior para el servidor; `go test ./...` y `go vet ./...`.
- Python 3.10 o posterior para `python3 -m unittest discover -s tools -p 'test_*.py'`.
- Node.js 22 o posterior: ejecuta `npm ci` y `npm test` tanto en
  `sdk/typescript` como en `examples/typescript-demo`.
- Las pruebas Go verifican también la documentación y las descargas de los SDKs.

Las pruebas no necesitan claves reales ni llamadas a Jev. Usa servidores falsos
o respuestas simuladas para cambios de integración. Conserva `.env`, claves y
ejemplos privados fuera de Git. Consulta `evals/README.es.md` para medir cambios en
el comportamiento semántico con casos previamente etiquetados.

Describe el caso de uso, el comportamiento esperado y una prueba que lo
reproduzca. Para reportar un problema de seguridad, no publiques claves ni
datos de clientes en un issue; contacta al administrador del despliegue.
