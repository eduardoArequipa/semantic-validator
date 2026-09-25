import { DirectValidator } from "@semantic-validator/sdk";
import { cases, describe, errorMessage, evaluate } from "./demo.js";

async function main() {
  const jevApiKey = process.env.TYPESAFE_API_KEY?.trim();
  if (!jevApiKey || jevApiKey === "pega_aqui_tu_clave_jev") {
    console.error("Configura TYPESAFE_API_KEY en .env con tu propia clave de Jev.");
    process.exitCode = 1;
    return;
  }
  const args = process.argv.slice(2);
  if (args.length && (args.length !== 2 || args[0] !== "--mensaje" || !args[1].trim())) {
    console.error('Uso: npm run prueba -- --mensaje "Texto de prueba sin datos personales"');
    process.exitCode = 1;
    return;
  }
  const selected = args.length
    ? [{ id: "personalizado", message: args[1], expected: "Compara el resultado con tu interpretación del mensaje." }]
    : cases;
  const client = new DirectValidator({ jevApiKey, timeoutMs: 40_000 });
  console.log(`Se harán ${selected.length * 2} consultas directamente a Jev con tu clave.`);
  console.log("Las expectativas son referencias humanas, no resultados garantizados. Confidence no es una garantía de acierto.");
  for (const item of selected) {
    console.log(`\n[${item.id}] ${item.message}\nReferencia: ${item.expected}`);
    const { purchase, complaint } = await evaluate(client, item.message);
    console.log(`Compra:  ${describe(purchase)}`);
    console.log(`Reclamo: ${describe(complaint)}`);
  }
}

main().catch(error => {
  console.error(errorMessage(error));
  console.error("Prueba detenida; algunas consultas pueden haberse cobrado en tu cuenta Jev. No compartas tu clave en reportes.");
  process.exitCode = 1;
});
