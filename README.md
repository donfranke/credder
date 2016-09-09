# Credder
Credential access API
# Description
Proof of concept credential management repository and service. To be used by apps and scripts to get credentials to protected resources like databases. Avoids hard-coding credentials.

# Dependencies
* Service is credder.go
* Sample client is client.py
* Can also use curl to test service
* Database is MongoDB
* D/encryption algorithm is Vigenere cipher (for testing purposes only)

# Usage
```
curl -H "Content-Type: application/json" -X POST -d "{\"Credid\":\"1\",\"Appname\":\"client.py\"}" http://localhost:8889/cred
curl -H "Content-Type: application/json" -X POST -d "{\"Keyid\":\"1\",\"Appname\":\"client.py\"}" http://localhost:8889/key
```
