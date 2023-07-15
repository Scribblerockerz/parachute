# 🪂 parachute

A backup utility for s3 compatible storages.

## Decrypt data with OpenSSL

Thanks to [go-openssl](https://github.com/Luzifer/go-openssl) it is possible to decrypt your data with openssl.

```bash
# decrypt data with OpenSSL 3.1.1 30 May 2023
openssl enc -d -aes-256-cbc -pbkdf2 -in archive.zip.enc -out your-data.zip
```
