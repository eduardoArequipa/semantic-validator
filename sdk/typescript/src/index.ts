export type ValidationStatus = "valid" | "invalid" | "uncertain";
export type BatchStatus = ValidationStatus | "error";

export interface ValidationResult {
  valid: boolean | null;
  confidence: number;
  status: ValidationStatus;
}

export interface BatchItem {
  id: string;
  rule: string;
  value: string;
}

export interface BatchItemError {
  code: string;
  message: string;
}

export interface BatchResult {
  id: string;
  valid: boolean | null;
  confidence: number;
  status: BatchStatus;
  error?: BatchItemError;
}

export interface ValidatorOptions {
  apiKey: string;
  baseURL?: string;
  timeoutMs?: number;
  fetcher?: typeof fetch;
}

export class SemanticValidatorError extends Error {
  readonly code: string;
  readonly statusCode?: number;

  constructor(message: string, options: { code?: string; statusCode?: number } = {}) {
    super(message);
    this.name = "SemanticValidatorError";
    this.code = options.code ?? "semantic_validator_error";
    this.statusCode = options.statusCode;
  }
}

export class Validator {
  private readonly apiKey: string;
  private readonly baseURL: string;
  private readonly timeoutMs: number;
  private readonly fetcher: typeof fetch;

  constructor(options: ValidatorOptions) {
    if (!options.apiKey?.trim()) {
      throw new TypeError("apiKey is required");
    }
    const timeoutMs = options.timeoutMs ?? 30_000;
    if (!Number.isFinite(timeoutMs) || timeoutMs <= 0) {
      throw new TypeError("timeoutMs must be a positive number");
    }

    this.apiKey = options.apiKey;
    this.baseURL = (options.baseURL ?? "http://localhost:8080").replace(/\/+$/, "");
    this.timeoutMs = timeoutMs;
    this.fetcher = options.fetcher ?? fetch;
  }

  async validate(rule: string, value: string): Promise<ValidationResult> {
    const payload = await this.post("/v1/validate", { rule, value });
    return parseValidationResult(payload);
  }

  name(value: string): Promise<ValidationResult> {
    return this.validate("person_name", value);
  }

  async check(value: string, question: string): Promise<ValidationResult> {
    const payload = await this.post("/v1/check", { value, question });
    return parseValidationResult(payload);
  }

  async validateBatch(items: readonly BatchItem[]): Promise<BatchResult[]> {
    if (items.length < 1 || items.length > 100) {
      throw new TypeError("items must contain between 1 and 100 elements");
    }

    const ids = new Set<string>();
    for (const item of items) {
      if (!item.id || !item.rule || typeof item.value !== "string") {
        throw new TypeError("each batch item requires an id, rule, and string value");
      }
      if (ids.has(item.id)) {
        throw new TypeError("batch item ids must be unique");
      }
      ids.add(item.id);
    }

    const payload = await this.post("/v1/validate/batch", { items });
    if (!isRecord(payload) || !Array.isArray(payload.items)) {
      throw new SemanticValidatorError("server returned an invalid batch response", {
        code: "invalid_response",
      });
    }
    return payload.items.map(parseBatchResult);
  }

  private async post(path: string, body: unknown): Promise<unknown> {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), this.timeoutMs);
    let response: Response;
    try {
      response = await this.fetcher(`${this.baseURL}${path}`, {
        method: "POST",
        headers: {
          Accept: "application/json",
          Authorization: `Bearer ${this.apiKey}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify(body),
        signal: controller.signal,
      });
    } catch (error) {
      if (controller.signal.aborted) {
        throw new SemanticValidatorError("request timed out", { code: "timeout" });
      }
      throw new SemanticValidatorError("request failed", { code: "transport_error" });
    } finally {
      clearTimeout(timeout);
    }

    let payload: unknown;
    try {
      payload = await response.json();
    } catch {
      throw new SemanticValidatorError("server returned invalid JSON", {
        code: "invalid_response",
        statusCode: response.status,
      });
    }

    if (!response.ok) {
      const error = isRecord(payload) && isRecord(payload.error) ? payload.error : {};
      throw new SemanticValidatorError(
        typeof error.message === "string" ? error.message : "request failed",
        {
          code: typeof error.code === "string" ? error.code : "http_error",
          statusCode: response.status,
        },
      );
    }
    return payload;
  }
}

function parseValidationResult(payload: unknown): ValidationResult {
  if (
    !isRecord(payload) ||
    !(typeof payload.valid === "boolean" || payload.valid === null) ||
    typeof payload.confidence !== "number" ||
    !isStatus(payload.status)
  ) {
    throw new SemanticValidatorError("server returned an invalid validation response", {
      code: "invalid_response",
    });
  }
  return {
    valid: payload.valid,
    confidence: payload.confidence,
    status: payload.status,
  };
}

function parseBatchResult(payload: unknown): BatchResult {
  if (
    !isRecord(payload) ||
    typeof payload.id !== "string" ||
    !(typeof payload.valid === "boolean" || payload.valid === null) ||
    typeof payload.confidence !== "number" ||
    !(isStatus(payload.status) || payload.status === "error")
  ) {
    throw new SemanticValidatorError("server returned an invalid batch item", {
      code: "invalid_response",
    });
  }
  if (payload.error !== undefined) {
    if (!isRecord(payload.error) || typeof payload.error.code !== "string" || typeof payload.error.message !== "string") {
      throw new SemanticValidatorError("server returned an invalid batch item error", {
        code: "invalid_response",
      });
    }
    return {
      id: payload.id,
      valid: payload.valid,
      confidence: payload.confidence,
      status: payload.status as BatchStatus,
      error: { code: payload.error.code, message: payload.error.message },
    };
  }
  if (payload.status === "error") {
    throw new SemanticValidatorError("batch error item is missing its error details", {
      code: "invalid_response",
    });
  }
  return {
    id: payload.id,
    valid: payload.valid,
    confidence: payload.confidence,
    status: payload.status,
  };
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function isStatus(value: unknown): value is ValidationStatus {
  return value === "valid" || value === "invalid" || value === "uncertain";
}

const JEV_URL = "https://api.typesafe.ai";
const NAME_QUESTION = "¿Este texto parece representar el nombre de una persona?";

export interface DirectValidatorOptions {
  jevApiKey: string;
  baseURL?: string;
  timeoutMs?: number;
  fetcher?: typeof fetch;
}

/** Sends requests directly to Jev; never contacts a Semantic Validator server. */
export class DirectValidator {
  private readonly apiKey: string;
  private readonly baseURL: string;
  private readonly timeoutMs: number;
  private readonly fetcher: typeof fetch;

  constructor(options: DirectValidatorOptions) {
    if (!options.jevApiKey?.trim()) throw new TypeError("jevApiKey is required");
    const timeoutMs = options.timeoutMs ?? 30_000;
    if (!Number.isFinite(timeoutMs) || timeoutMs <= 0) throw new TypeError("timeoutMs must be positive");
    const url = new URL(options.baseURL ?? JEV_URL);
    if (!(url.protocol === "https:" && url.hostname === "api.typesafe.ai") &&
        !(url.protocol === "http:" && ["localhost", "127.0.0.1", "[::1]"].includes(url.hostname))) {
      throw new TypeError("Jev endpoint must be api.typesafe.ai (except loopback tests)");
    }
    if (url.username || url.password || url.search || url.hash || url.pathname !== "/") throw new TypeError("invalid Jev endpoint");
    this.apiKey = options.jevApiKey.trim();
    this.baseURL = url.href.replace(/\/+$/, "");
    this.timeoutMs = timeoutMs;
    this.fetcher = options.fetcher ?? fetch;
  }

  validate(rule: string, value: string): Promise<ValidationResult> {
    validDirectText(value, "value");
    if (rule !== "person_name") throw new SemanticValidatorError("requested rule is not registered", { code: "unknown_rule" });
    return this.check(value, NAME_QUESTION);
  }

  name(value: string): Promise<ValidationResult> {
    return this.validate("person_name", value);
  }

  async check(value: string, question: string): Promise<ValidationResult> {
    const state = validDirectText(value, "value");
    const instructions = validDirectText(question, "question");
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), this.timeoutMs);
    let response: Response;
    try {
      response = await this.fetcher(`${this.baseURL}/v1/systemone`, {
        method: "POST",
        headers: { Accept: "application/json", Authorization: `Bearer ${this.apiKey}`, "Content-Type": "application/json" },
        body: JSON.stringify({ model: "jev-latest", state, questions: { result: {
          type: "choice", instructions, criteria: { true: "El texto cumple la condición.", false: "El texto no cumple la condición." },
        } } }),
        redirect: "error",
        signal: controller.signal,
      });
    } catch {
      throw new SemanticValidatorError(controller.signal.aborted ? "Jev request timed out" : "Jev request failed", {
        code: controller.signal.aborted ? "provider_timeout" : "provider_error",
      });
    } finally {
      clearTimeout(timer);
    }
    if (!response.ok) throw new SemanticValidatorError(`Jev returned HTTP ${response.status}`, { code: "provider_error", statusCode: response.status });
    let payload: unknown;
    try { payload = await response.json(); }
    catch { throw new SemanticValidatorError("Jev returned invalid JSON", { code: "invalid_response" }); }
    const answer = isRecord(payload) && isRecord(payload.answers) ? payload.answers.result : undefined;
    if (!isRecord(answer) || !["true", "false"].includes(answer.choice as string)) {
      throw new SemanticValidatorError("Jev returned an invalid choice", { code: "invalid_response" });
    }
    const confidence = typeof answer.confidence === "number" ? answer.confidence :
      isRecord(answer.probabilities) ? answer.probabilities[answer.choice as string] : undefined;
    if (typeof confidence !== "number" || !Number.isFinite(confidence) || confidence < 0 || confidence > 1) {
      throw new SemanticValidatorError("Jev returned an invalid confidence", { code: "invalid_response" });
    }
    if (confidence < 0.8) return { valid: null, confidence, status: "uncertain" };
    const valid = answer.choice === "true";
    return { valid, confidence, status: valid ? "valid" : "invalid" };
  }

  async validateBatch(items: readonly BatchItem[]): Promise<BatchResult[]> {
    if (items.length < 1 || items.length > 100) throw new TypeError("items must contain between 1 and 100 elements");
    const ids = new Set<string>();
    for (const item of items) {
      if (!item || !item.id?.trim() || typeof item.rule !== "string" || typeof item.value !== "string" || ids.has(item.id)) {
        throw new TypeError("batch items require unique ids, a rule, and a string value");
      }
      ids.add(item.id);
    }
    const results = new Array<BatchResult>(items.length);
    let next = 0;
    await Promise.all(Array.from({ length: Math.min(4, items.length) }, async () => {
      while (next < items.length) {
        const index = next++;
        const item = items[index]!;
        try {
          results[index] = { id: item.id, ...await this.validate(item.rule, item.value) };
        } catch (error) {
          const code = error instanceof SemanticValidatorError ? error.code : "provider_error";
          results[index] = { id: item.id, valid: null, confidence: 0, status: "error", error: {
            code, message: code === "unknown_rule" ? "requested rule is not registered" : code === "invalid_request" ? "invalid value" : "semantic provider request failed",
          } };
        }
      }
    }));
    return results;
  }
}

function validDirectText(value: string, field: string): string {
  if (typeof value !== "string" || !value.trim() || Array.from(value.trim()).length > 10_000) {
    throw new SemanticValidatorError(`${field} must contain between 1 and 10000 characters`, { code: "invalid_request" });
  }
  return value.trim();
}
