import json
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
CATALOG = ROOT / "internal/rules/catalog.json"
DIRECT_CLIENTS = (
    ROOT / "sdk/typescript/src/index.ts",
    ROOT / "sdk/python/src/semantic_validator/client.py",
    ROOT / "sdk/go/direct.go",
    ROOT / "sdk/java/src/main/java/io/semanticvalidator/sdk/DirectValidator.java",
)


class RuleCatalogTests(unittest.TestCase):
    def test_unique_versioned_rules_and_direct_client_parity(self):
        rules = json.loads(CATALOG.read_text(encoding="utf-8"))
        self.assertEqual({rule["id"] for rule in rules},
                         {"person_name", "product_description", "address"})
        self.assertEqual(len(rules), 3)
        for rule in rules:
            self.assertRegex(rule["id"], r"^[a-z][a-z_]*$")
            self.assertGreaterEqual(rule["version"], 1)
            self.assertTrue(rule["question"].strip())
            self.assertTrue(rule["description"].strip())
            for path in DIRECT_CLIENTS:
                source = path.read_text(encoding="utf-8")
                self.assertIn(rule["id"], source, path.name)
                self.assertIn(rule["question"], source, path.name)


if __name__ == "__main__":
    unittest.main()
