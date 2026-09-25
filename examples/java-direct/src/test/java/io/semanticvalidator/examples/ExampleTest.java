package io.semanticvalidator.examples;

import com.sun.net.httpserver.HttpServer;
import io.semanticvalidator.sdk.DirectValidator;
import org.junit.jupiter.api.Test;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.PrintStream;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.time.Duration;
import java.util.concurrent.atomic.AtomicInteger;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

class ExampleTest {
    @Test
    void runsWithoutARealKey() throws Exception {
        AtomicInteger calls = new AtomicInteger();
        HttpServer server = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
        server.createContext("/v1/systemone", exchange -> {
            calls.incrementAndGet();
            if (!"Bearer test-only".equals(exchange.getRequestHeaders().getFirst("Authorization"))) {
                exchange.sendResponseHeaders(401, -1);
                exchange.close();
                return;
            }
            String request = new String(exchange.getRequestBody().readAllBytes(), StandardCharsets.UTF_8);
            double confidence = request.contains("nombre de una persona") ? 0.96 : 0.62;
            byte[] response = ("{\"answers\":{\"result\":{\"choice\":\"true\",\"confidence\":"
                    + confidence + "}}}").getBytes(StandardCharsets.UTF_8);
            exchange.getResponseHeaders().set("Content-Type", "application/json");
            exchange.sendResponseHeaders(200, response.length);
            exchange.getResponseBody().write(response);
            exchange.close();
        });
        server.start();
        try {
            DirectValidator validator = new DirectValidator("test-only",
                    "http://127.0.0.1:" + server.getAddress().getPort(), Duration.ofSeconds(5));
            ByteArrayOutputStream output = new ByteArrayOutputStream();
            Example.run(validator, new PrintStream(output, true, StandardCharsets.UTF_8));
            String printed = output.toString(StandardCharsets.UTF_8);
            assertEquals(2, calls.get());
            assertTrue(printed.contains("nombre: true"));
            assertTrue(printed.contains("reclamo: requiere revisión"));
        } finally {
            server.stop(0);
        }
    }

    @Test
    void providerFailureIsNotFalse() throws Exception {
        HttpServer server = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
        server.createContext("/v1/systemone", exchange -> {
            exchange.sendResponseHeaders(503, -1);
            exchange.close();
        });
        server.start();
        try {
            DirectValidator validator = new DirectValidator("test-only",
                    "http://127.0.0.1:" + server.getAddress().getPort(), Duration.ofSeconds(5));
            assertThrows(IOException.class, () -> Example.run(validator, System.out));
        } finally {
            server.stop(0);
        }
    }
}
