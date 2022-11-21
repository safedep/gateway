# Development

## PDP Development

Build and run the PDP using:

```bash
cd services && make
GLOBAL_CONFIG_PATH=../config/gateway.json PDP_POLICY_PATH=../policies ./out/pdp-server
```

PDP listens on `0.0.0.0:9000`. To use the host instance of PDP, edit `config/envoy.json` and set the address of the `ExtAuthZ` plugin to your host network address.

## Policy Development

Policies are written in [Rego](https://www.openpolicyagent.org/docs/latest/policy-language/) and evaluated with [Open Policy Agent](https://www.openpolicyagent.org/docs/latest/integration/#integrating-with-the-go-api)

To run policy test cases:

```bash
cd policies && make test
```

* Refer to `policies/example.rego` for policy example
* Policies are load from `./policies` directory

## Tap Development

The *Tap Service* is integrated as a Envoy [ExtProc](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_proc_filter) filter. This means, it has greater control over Envoy's request processing life-cycle and can make changes if required.

Currently, it is used for publishing events for data collection only but in future may be extended to support other use-cases. Tap service internally implements a handler chain to delegate an Envoy event to its internal handlers. Example:

```go
tapService, err := tap.NewTapService(config, []tap.TapHandlerRegistration{
  tap.NewTapEventPublisherRegistration(),
})
```

To build and use from host:

```bash
cd services && make
GLOBAL_CONFIG_PATH=../config/gateway.json ./out/tap-server
```

> To use Tap service from host, edit `envoy.json` and change address of `ext-proc-tap` cluster.

## Debug NATS Messaging

Start a docker container with `nats` client

```bash
docker run --rm -it \
   --network supply-chain-security-gateway_default \
   -v `pwd`:/workspace \
   synadia/nats-box
```

Subscribe to a subject and receive messages

```bash
GODEBUG=x509ignoreCN=0 nats sub \
   --tlscert=/workspace/pki/tap/server.crt \
   --tlskey=/workspace/pki/tap/server.key \
   --tlsca=/workspace/pki/root.crt \
   --server=tls://nats-server:4222 \
   com.msg.event.upstream.request
```


