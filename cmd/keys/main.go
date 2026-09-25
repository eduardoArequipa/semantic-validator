// keys administers the same local store as the server. Tokens print only on creation.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/eduardoArequipa/semantic-validator/internal/access"
	"log"
	"os"
	"sort"
)

func main() {
	path := flag.String("store", os.Getenv("API_KEYS_FILE"), "persistent API key store path")
	name := flag.String("name", "", "tester or project name")
	limit := flag.Int("daily-limit", 100, "validation units per UTC day")
	id := flag.String("id", "", "key ID to revoke")
	flag.Parse()
	if *path == "" || flag.NArg() != 1 {
		log.Fatal("usage: keys -store data/keys.json [-name NAME -daily-limit 100 | -id ID] create|list|revoke")
	}
	switch flag.Arg(0) {
	case "create", "list", "revoke":
	default:
		log.Fatal("unknown command")
	}
	store, err := access.Open(*path)
	if err != nil {
		log.Fatal(err)
	}
	switch flag.Arg(0) {
	case "create":
		keyID, token, err := store.Create(*name, *limit)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ID: %s\nAPI key (shown once): %s\n", keyID, token)
	case "revoke":
		if err := store.Revoke(*id); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Key revoked:", *id)
	case "list":
		keys, err := store.List()
		if err != nil {
			log.Fatal(err)
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i].ID < keys[j].ID })
		if err := json.NewEncoder(os.Stdout).Encode(keys); err != nil {
			log.Fatal(err)
		}
	}
}
