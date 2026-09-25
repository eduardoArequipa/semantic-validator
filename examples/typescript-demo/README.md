# Prueba directa con Jev: compra y reclamos

Este ejemplo usa el SDK TypeScript `0.3.0` desde Node.js 22+. Analiza cada mensaje con dos preguntas: intención de compra y presentación de un reclamo. Las decisiones pueden ser sí, no o inciertas.

Necesitas tu **propia clave de Jev** de TypeSafe. No necesitas nuestro servidor ni una clave de Semantic Validator. El SDK está incluido en `vendor/`; `npm ci` descarga las herramientas de compilación fijadas en el archivo de bloqueo.

```bash
npm ci
cp .env.example .env
# Edita .env y configura TYPESAFE_API_KEY con tu propia clave.
npm run prueba
```

La ejecución normal realiza diez consultas a Jev: dos por cada uno de los cinco mensajes ficticios. Cada consulta puede consumir uso de tu cuenta. Para probar un solo mensaje:

```bash
npm run prueba -- --mensaje "Quiero comprar dos unidades, pero mi pedido anterior llegó roto."
```

`npm run prueba` carga `.env` en Node; no copies la clave al código, no compartas `.env` y no uses esta librería con claves privadas en el navegador. El texto se envía directamente a Jev, no a `validator.trialsur.cloud`. La API alojada permanece desactivada.

Una confianza inferior a 0,80 produce `status=uncertain` y `valid=null`; necesita revisión humana. La confianza no es una garantía de corrección. El ejemplo imprime el mensaje en tu consola, pero no guarda un archivo de resultados. Usa solo mensajes ficticios.

Si Jev devuelve 401, revisa `TYPESAFE_API_KEY`. Si devuelve 429, revisa los límites de tu cuenta y espera antes de repetir. El ejemplo no reintenta automáticamente, porque una consulta ya puede haber consumido uso.

`npm test` usa respuestas simuladas y no necesita una clave ni acceso a Jev. Los detalles de implementación están en `src/prueba.ts` y `src/demo.ts`.

## Formulario semántico

El ejemplo `src/form.ts` valida nombre, descripción de producto y dirección
con tres reglas versionadas. Devuelve `accepted`, `rejected`, `review` o `error`
por campo: un fallo de Jev nunca se interpreta como campo inválido. Ejecuta
`npm test` sin clave para ver las cuatro rutas con respuestas simuladas. Para
probarlo en vivo con tu propia clave en `.env`, ejecuta `npm run form` desde
esta carpeta. Se harán hasta tres consultas a Jev y no se imprime el texto
del formulario. Úsalo solo en un backend, nunca en código del navegador.
