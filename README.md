# 🪂 parachute

A backup utility for s3 compatible storages.

## Install

TBD

## Usage

```sh
# Zip and upload files to an s3 bucket
parachute backup ./uploads/* --remote s3://some-bucket/uploads.zip

# Encrypt the data before upload
parachute backup ./uploads/* --pass s3cr3t --remote s3://some-bucket/uploads.zip.enc

# Download and unzip a remote target
parachute restore ./downloads --remote s3://some-bucket/uploads.zip

# Download and unzip a remote target (previously encrypted)
parachute restore ./downloads --pass s3cr3t --remote s3://some-bucket/uploads.zip.enc

# Zip files into an archive, an place it  somewhere`./backups/20060102150405_archive.zip.enc`
parachute pack ./uploads/* --pass s3cr3t --output ./backups/ --timed-name

# Unzip an (encrypted) archive, and place it somewhere`./somewhere`
parachute unpack 20060102150405_archive.zip.enc --pass s3cr3t --output ./somewhere
```

## Decrypt data with OpenSSL

Thanks to [go-openssl](https://github.com/Luzifer/go-openssl) it is possible to decrypt your data with openssl.

```sh
# decrypt data with OpenSSL 3.1.1 30 May 2023
openssl enc -d -aes-256-cbc -pbkdf2 -in archive.zip.enc -out your-data.zip
```

## Configuration

### parachute.toml

Configuration is loaded from the following places:

```
/etc/parachute/parachute.toml
$HOME/.config/parachute/parachute.toml
./parachute.toml
```

Following options can be configured:

```toml
# available log levels (debug, info, warn, error, fatal, panic)
log_level = "error"

# available log formats (console/json)
log_format = "json"

# encrypt an archive with given passphrase
passphrase = "some-fancy-passphrase"

# prevent encryption
no_encryption = false

# prefix current date/time when running `pack`
timed_name = false

# S3 endpoint and access
endpoint = ""
access_key = ""
secret_key = ""

# remote archive destination, .enc for encrypted targets
remote = "s3://bucket-name/file-name.zip.enc"
```

### Environment

All of the options can be overwritten by env vars. It's uppercased version, overwrite the file configuration.

```sh
export PARACHUTE_ENDPOINT=s3.eu-central-2.wasabisys.com
export PARACHUTE_ACCESS_KEY=access-key-goes-here
export PARACHUTE_SECRET_KEY=secret-key-goes-here
export PARACHUTE_PASSPHRASE=cilk-and-mookies
export PARACHUTE_LOG_LEVEL=debug
export PARACHUTE_REMOTE=s3://bucket-for-backups/assets.zip.enc
```
