# credder
Credential access API
# Usage
```
curl -H "Content-Type: application/json" -X POST -d "{\"Credid\":\"1\",\"Appname\":\"client.py\"}" http://localhost:8889/cred
curl -H "Content-Type: application/json" -X POST -d "{\"Keyid\":\"1\",\"Appname\":\"client.py\"}" http://localhost:8889/key
```