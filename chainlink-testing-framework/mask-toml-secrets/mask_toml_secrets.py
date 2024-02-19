import toml
import base64
import sys
import unittest

def find_secret_values(obj):
    secret_values = []

    if isinstance(obj, dict):
        for key, value in obj.items():
            if isinstance(value, (dict, list)):
                if key.endswith("_secret") and isinstance(value, list):
                    # Directly add all items in the list if the key ends with '_secret'
                    secret_values.extend(value)
                else:
                    # Recursively search within the dict or list
                    secret_values.extend(find_secret_values(value))
            elif key.endswith("_secret"):
                secret_values.append(value)
    elif isinstance(obj, list):
        for item in obj:
            secret_values.extend(find_secret_values(item))

    return secret_values

# Main execution
def main():
    if len(sys.argv) < 2:
        print("Usage: python mask_toml_secrets.py <base64_toml1> <base64_toml2> ...")
        sys.exit(1)

    print(f"Provided {len(sys.argv)-1} base64_toml")

    all_secrets = []
    for encoded_data in sys.argv[1:]:
        try:
            decoded_data = base64.b64decode(encoded_data).decode("utf-8")
            data = toml.loads(decoded_data)
            if not data:  # Check if the data is empty
                raise ValueError("TOML config is empty")
        except Exception as e:
            raise Exception(f"Could not decode TOML config. Error: {e}") from None            
        secrets = find_secret_values(data)
        all_secrets.extend(secrets)

    for secret in all_secrets:
        print(f"::add-mask::{secret}")

    print(f"Masked {len(all_secrets)} secrets")

# Test cases
class TestSecretValueFinder(unittest.TestCase):
    def test_find_secret_values(self):
        tests = [
            {
                "name": "Single secret",
                "input": {"api_secret": "12345"},
                "expectedSecrets": ["12345"],
            },
            {
                "name": "Multiple secrets",
                "input": {
                    "api_secret": "12345",
                    "nested": {
                        "db_secret": "abcde",
                        "nested_list": [
                            {"1_secret": "abc"},
                            {"second_secret": "def"},
                            {"api_key": "g"},
                        ],
                    },
                },
                "expectedSecrets": ["12345", "abcde", "abc", "def"],
            },
            {
                "name": "Multiple mixed type secrets",
                "input": {
                    "string_secret": "secret",
                    "int_secret": 123,
                    "bool_secret": True,
                    "float_secret": 3.14,
                    "nested": {"nested_secret": "nested"},
                },
                "expectedSecrets": ["secret", "123", "True", "3.14", "nested"],
            },
                        {
                "name": "List secrets",
                "input": {
                    "nested": {"nested_secret": ["abc", True, 3.14]},
                },
                "expectedSecrets": ["3.14", "True", "abc"],
            },
            {
                "name": "No secret",
                "input": {"api_key": "abcde"},
                "expectedSecrets": [],
            },
        ]

        for test in tests:
            with self.subTest(name=test["name"]):
                result = find_secret_values(test["input"])
                self.assertEqual(sorted(map(str, result)), sorted(map(str, test["expectedSecrets"])))

if __name__ == '__main__':
    if '--test' in sys.argv:
        sys.argv.remove('--test')  # Remove the test argument before running unittest
        unittest.main()
    else:
        main()
