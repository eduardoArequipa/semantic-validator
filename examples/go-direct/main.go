package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	validator "github.com/eduardoArequipa/semantic-validator/sdk/go"
)

func main() {
	key := os.Getenv("TYPESAFE_API_KEY")
	if key == "" {
		log.Fatal("TYPESAFE_API_KEY is required; try go test ./examples/go-direct first")
	}
	client, err := validator.NewDirectClient(key)
	if err != nil {
		log.Fatal(err)
	}
	if err := run(context.Background(), client, os.Stdout); err != nil {
		log.Fatalf("Jev request failed (not an invalid result): %v", err)
	}
}

func run(ctx context.Context, client *validator.DirectClient, out io.Writer) error {
	name, err := client.Name(ctx, "Jorge Eduardo")
	if err != nil {
		return fmt.Errorf("name check: %w", err)
	}
	if err := show(out, "nombre", name); err != nil {
		return err
	}
	complaint, err := client.Check(ctx,
		"Mi pedido llegó dañado y quiero una solución",
		"¿El cliente está presentando un reclamo?")
	if err != nil {
		return fmt.Errorf("complaint check: %w", err)
	}
	return show(out, "reclamo", complaint)
}

func show(out io.Writer, label string, result validator.Result) error {
	if result.Status == "uncertain" {
		_, err := fmt.Fprintf(out, "%s: requiere revisión (confianza %.2f)\n", label, result.Confidence)
		return err
	}
	if result.Valid == nil {
		return fmt.Errorf("%s: missing decision", label)
	}
	_, err := fmt.Fprintf(out, "%s: %t (confianza %.2f)\n", label, *result.Valid, result.Confidence)
	return err
}
