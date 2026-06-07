# Internal CA certificates

Place internal corporate CA certificates in this directory with a `.crt` extension, for example:

```text
certs/internal-ca.crt
```

The Docker build copies this directory into the build image before dependency download and merges any `.crt` files into the container CA bundle. The final runtime image receives the generated CA bundle, but it does not receive HTTP proxy environment variables.

Do not commit private keys or secret material here; commit only public CA certificates that are approved for distribution with the application image.
