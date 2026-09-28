# Fitting Progressive API to WAF

---
## Layer 1: Zero Configuration

User only needs to provide the minimum configuration and the cloud manager will handle the rest using defaults.
The AppLoadBalancer CR defines the configuration for the application load balancer, which is required for the WAF to function properly.

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: AppLoadBalancer
metadata:
  name: my-app
spec:
  backend:
    apiVersion: v1
    kind: Service
    name: my-service
    namespace: default
```
The example above is the minimum configuration required to create an AppLoadBalancer CR. The backend field specifies the service that the load balancer will route traffic to.
With this configuration, cloud manager will:
 - automatically detect the cloud provider,
 - since there is no TLS configuration, it will create an HTTP load balancer on port 80,
 - since there is no Hostname configuration, it will create a load balancer using a wildcard hostname,
 - since there is no Health check configuration, it will create a default health check on / using the service default port.

Depending on customer requirements, it is also possible to apply to this load balancer the default WAF configuration, which will create a WAF CRD with the default configuration.

## Industry Patterns:

### Common Base with Provider-Specific Extensions
The user provides a base configuration, and the cloud manager generates provider-specific resources based on that configuration, it also allows for extensions on the backend.
This example shows a Service, but it could be an istio gateway, and cloud manager will fetch all required data from it.

### Normalized Status Contract
The cloud manager will provide a normalized status contract for the user, regardless of the underlying cloud provider.

```yaml
status:
  certificates:
    - source: some-names/my-other-tls-secret
      observedGeneration: 12
      hash: alkfhkjafhkjahf
      providerId: asjhaljfhakjsfhdkj
      status: Processing
      message: Error creating certificate
      lastTransitionTime: 2026-01-01T00:00:00.0000Z
  observedGeneration: 34
  conditions:
    - type: Ready
      status: Unknown
      reason: Processsing
```
---
## Layer 2: Custom Configuration

User can provide custom configuration to override the defaults set by the cloud manager. This allows for more control over the behavior of the load balancer and WAF.

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: AppLoadBalancer
metadata:
    name: my-app
spec:
    backend:
        apiVersion: v1
        kind: Service
        name: my-service
        namespace: default
        healthCheck:
          path: /healthz
          nodePort: 32451
          timeout: 5s
          retryCount: 2
    frontend:
      - hosts:
          - my-service.example.com
        port:
          number: 80
          protocol: HTTP
        httpsRedirect: true
      - hosts:
          - my-service.example.com
        port:
          number: 443
          protocol: HTTPS
        tls:
          secretRef:
            name: my-tls-secret
            namespace: some-namespace
  # Optional, LoadBalancer can be created and functional without policy
    policy:
      apiVersion: cloud-resources.kyma-project.io/v1alpha1
      kind: WafPolicy
      name: my-gcp-policy
```

In this example, the user has provided a custom health check configuration for the backend service, as well as a custom frontend configuration that includes two hosts with different ports and protocols. The user has also specified a TLS secret for the HTTPS host.
The first iteraction of WAF will support only a limited set of features (OWASP top 10 applied by default), but it is possible to extend the configuration in the future to support more advanced features. The goal is to provide a simple and easy-to-use interface for the user, while still allowing for more advanced configuration options for those who need them.

---
## Layer 3: Advanced Configuration

Cloud manager will generate a provider-specific WAF configuration. The user can change it to customize the behavior of the load balancer and WAF even further to better suit their needs. This allows for more fine-grained control over the behavior of the load balancer and WAF.


```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafPolicy
metadata:
  name: my-gcp-policy
spec:
  data: |
  {
    "rules": [
      {
        "priority": 1000,
        "action": "deny(403)",
        "description": "Block all OWASP core threats at sensitivity 1",
        "preview": false,
        "match": {
          "expr": {
            "expression": "evaluatePreconfiguredWaf('sqli-v422-stable', {'sensitivity': 1}) || evaluatePreconfiguredWaf('xss-v422-stable', {'sensitivity': 1}) || evaluatePreconfiguredWaf('lfi-v422-stable', {'sensitivity': 1}) || evaluatePreconfiguredWaf('rfi-v422-stable', {'sensitivity': 1}) || evaluatePreconfiguredWaf('rce-v422-stable', {'sensitivity': 1}) || evaluatePreconfiguredWaf('methodenforcement-v422-stable', {'sensitivity': 1}) || evaluatePreconfiguredWaf('scannerdetection-v422-stable', {'sensitivity': 1}) || evaluatePreconfiguredWaf('protocolattack-v422-stable', {'sensitivity': 1}) || evaluatePreconfiguredWaf('php-v422-stable', {'sensitivity': 1}) || evaluatePreconfiguredWaf('sessionfixation-v422-stable', {'sensitivity': 1})"
          }
        }
      }
    ]
  }
```

## Industry Patterns:

### Common Base with Provider-Specific Extensions
Cloud manager will generate a provider-specific WAF configuration based on the user's custom configuration. The user can then modify the generated configuration to further customize the behavior of the load balancer and WAF. This allows for more fine-grained control over the behavior of the load balancer and WAF, while still providing a simple and easy-to-use interface for the user.

### Normalized Status Contract
The cloud manager will provide a normalized status contract for the user, regardless of the underlying cloud provider.

```yaml
status:
  observedGeneration: 9
  conditions:
    - type: Ready
      status: False
      reason: Ready
  providerId: aksjdhakjsdh
```