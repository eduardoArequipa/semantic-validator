package io.semanticvalidator.sdk;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.sun.net.httpserver.HttpServer;
import org.junit.jupiter.api.Test;

import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.time.Duration;
import java.util.List;
import java.util.concurrent.atomic.AtomicInteger;

import static org.junit.jupiter.api.Assertions.*;

class DirectValidatorTest {
    @Test
    void directClientAndBatch() throws Exception {
        HttpServer server = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
        AtomicInteger calls = new AtomicInteger();
        server.createContext("/v1/systemone", exchange -> {
            calls.incrementAndGet();
            assertEquals("Bearer own-key", exchange.getRequestHeaders().getFirst("Authorization"));
            JsonNode body = new ObjectMapper().readTree(exchange.getRequestBody());
            assertEquals("jev-latest", body.path("model").asText());
            assertEquals("choice", body.path("questions").path("result").path("type").asText());
            String state = body.path("state").asText();
            String choice = "No".equals(state) ? "false" : "true";
            String confidence = "No".equals(state) ? "0.98" : "Maybe".equals(state) ? "0.63" : "0.93";
            byte[] answer = ("{\"answers\":{\"result\":{\"choice\":\"" + choice + "\",\"confidence\":" + confidence + "}}}").getBytes(StandardCharsets.UTF_8);
            exchange.sendResponseHeaders(200, answer.length);
            try (var output = exchange.getResponseBody()) { output.write(answer); }
        });
        server.start();
        try {
            DirectValidator client = new DirectValidator("own-key", "http://127.0.0.1:" + server.getAddress().getPort(), Duration.ofSeconds(10));
            ValidationResult result = client.name("Jorge Eduardo");
            assertEquals(Boolean.TRUE, result.getValid());
            assertEquals("valid", result.getStatus());
            assertEquals(Boolean.TRUE, client.validate("product_description", "Taladro de 18 V").getValid());
            assertEquals(Boolean.TRUE, client.validate("address", "Av. Bolívar 123").getValid());
            assertEquals(Boolean.FALSE, client.check("No", "¿Compra?").getValid());
            ValidationResult uncertain = client.check("Maybe", "Question");
            assertNull(uncertain.getValid());
            assertEquals("uncertain", uncertain.getStatus());
            List<BatchResult> batch = client.validateBatch(List.of(
                    new BatchItem("a", "person_name", "Jorge"),
                    new BatchItem("b", "missing", "No"),
                    new BatchItem("c", "person_name", " ")));
            assertEquals(List.of("a", "b", "c"), List.of(batch.get(0).getId(), batch.get(1).getId(), batch.get(2).getId()));
            assertEquals("unknown_rule", batch.get(1).getError().getCode());
            assertEquals("invalid_request", batch.get(2).getError().getCode());
            assertEquals(6, calls.get());
        } finally {
            server.stop(0);
        }
    }

    @Test
    void rejectsMissingKeyAndInsecureEndpoint() {
        assertThrows(IllegalArgumentException.class, () -> new DirectValidator(""));
        assertThrows(IllegalArgumentException.class, () -> new DirectValidator("key", "http://example.com", Duration.ofSeconds(1)));
    }
}
