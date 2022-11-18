# Authentication

There are two authentication points:

1. Ingress
2. Egress

Ingress authentication is for incoming requests to the gateway and can be used to identify who is accessing the gateway.

Egress authentication is for upstream repositories, especially the ones that need authentication e.g. CodeArtifact, JFrog, Nexus etc.

- [Ingress Gateway Authentication](Gateway-Authentication.md)

#### Ingress Authentication

##### Basic Authentication

Use `htpasswd` to add users:

```bash
htpasswd -nbB user1 password1 >> ./config/gateway-auth-basic
```

Enable authentication for upstream in `config/global.yml`


