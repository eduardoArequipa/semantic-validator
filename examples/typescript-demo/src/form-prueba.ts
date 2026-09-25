import { DirectValidator } from "@semantic-validator/sdk";
import { validateForm } from "./form.js";

async function main() {
  const jevApiKey = process.env.TYPESAFE_API_KEY?.trim();
  if (!jevApiKey || jevApiKey === "pega_aqui_tu_clave_jev") {
    console.error("Configura tu propia TYPESAFE_API_KEY en .env; no la incluyas en el código.");
    process.exitCode = 1;
    return;
  }
  const validator = new DirectValidator({ jevApiKey });
  console.log("Validando tres campos ficticios directamente con Jev; pueden consumir tres consultas.");
  const decisions = await validateForm(validator, {
    name: "Ana Lucía Pérez",
    productDescription: "Taladro inalámbrico de 18 V con dos baterías",
    address: "Av. Libertad 123, piso 2",
  });
  for (const [field, result] of Object.entries(decisions)) {
    console.log(`${field}: ${result.state} (${result.confidence ?? "sin confianza"}) — ${result.message}`);
  }
}

main().catch(() => {
  console.error("No se pudo completar la validación; algunas consultas pueden haberse cobrado en tu cuenta Jev.");
  process.exitCode = 1;
});
