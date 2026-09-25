# Claves individuales y cuotas

El modo individual usa un archivo persistente en un volumen local del servidor Linux.
Permite administrar claves sin reiniciar el servicio. No requiere PostgreSQL. Está
diseñado para un solo host y un grupo pequeño de testers; no usar sobre NFS ni repartir
el archivo entre hosts. Cada operación actualiza el archivo bajo un bloqueo entre
procesos. El volumen contiene hashes de las claves, nombres y contadores, no mensajes.

## Activar en Docker

Desde la carpeta del proyecto, construir y arrancar la nueva imagen manteniendo
todavía la configuración actual:

```bash
docker compose -f docker-compose.prod.yml up --build -d
docker compose -f docker-compose.prod.yml exec semantic-validator /app/keys -store /data/keys.json -name tester-jorge -daily-limit 100 create
```

Guardar la clave mostrada y entregarla únicamente a su destinatario: no se puede
recuperar después. Repetir el comando con otro nombre para cada tester.

En el `.env` del servidor, añadir o modificar:

```dotenv
API_KEYS_FILE=/data/keys.json
```

Activar con:

```bash
docker compose -f docker-compose.prod.yml up -d
```

Al activar este modo, `SEMANTIC_VALIDATOR_API_KEY` deja de aceptarse, incluso si
permanece en el entorno. Cambiar los clientes a sus claves individuales. La clave
de Jev (`TYPESAFE_API_KEY`) continúa exclusivamente en el backend. Sin API_KEYS_FILE
el servidor mantiene el modo anterior, sin cuotas individuales.

## Administrar

```bash
docker compose -f docker-compose.prod.yml exec semantic-validator /app/keys -store /data/keys.json list
docker compose -f docker-compose.prod.yml exec semantic-validator /app/keys -store /data/keys.json -id ID_DE_LA_CLAVE revoke
```

La lista muestra ID, nombre, límite diario, estado de revocación, día UTC, unidades
consumidas (`used`) e intentos de inferencia (`provider_attempts`); no muestra tokens.
Revocar impide nuevas peticiones y nuevas llamadas al proveedor; una llamada que ya
está en curso puede terminar. Para rotar una clave, crear otra y revocar la anterior.

Para desarrollo local con Go, usar `go run ./cmd/keys -store data/keys.json ...`
y `API_KEYS_FILE=data/keys.json` al iniciar el servidor. Go no carga `.env` por sí solo.

## Cuotas

- Reinicio diario a las 00:00 UTC, sin tareas programadas.
- Validate y check reservan una unidad después de decodificar la solicitud; validate
  además exige un identificador de regla no vacío antes de reservar.
- Batch reserva todas sus unidades tras validar tamaño e IDs. Si no alcanza la cuota,
  responde HTTP 429 con `quota_exceeded`, sin procesar elementos ni consumir cuota.
- La caché, valores inválidos, reglas desconocidas y fallos de Jev consumen unidades
  reservadas. JSON inválido, método incorrecto y batch estructuralmente inválido no.
- Las reservas no se reembolsan al desconectarse el cliente o fallar el proveedor.
- `provider_attempts` cuenta intentos persistidos antes de llamar a Jev, incluso fallos.
  Si el proceso cae entre el registro y el envío, puede contar un intento no enviado;
  no es un registro de facturación del proveedor.
- HTTP 429 incluye `Retry-After`; las reservas aceptadas incluyen `X-Quota-Limit` y
  `X-Quota-Remaining`. El límite existente de 60 solicitudes/minuto por IP sigue vigente.
- Al fallar el almacenamiento se bloquean nuevas operaciones (HTTP 503). Si falla
  durante un batch, los elementos afectados tienen `usage_unavailable`.

El volumen `access-data` conserva el consumo y las claves tras recrear el contenedor.
No eliminarlo con `docker compose down -v`. Incluirlo en los backups; restaurar un
backup antiguo también restaura sus contadores y estados de revocación.
