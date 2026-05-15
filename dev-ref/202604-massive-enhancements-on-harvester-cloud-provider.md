# Massive Enhancements on Harvester Cloud Provider

## Motivations

The primary goal of these enhancements is to transition the Harvester Cloud Provider (HCP) from an **automatic discovery (best-effort)** model to a **deterministic configuration** model. 

In complex environments—such as those involving multi-network configurations or multiple IP assignments—the HCP requires explicit guidance to correctly identify management networks and internal IPs. This ensures stable cluster communication and predictable control plane behavior.

EPIC: https://github.com/harvester/harvester/issues/10068

---

## Summary of Configuration Flags

| Pillar | Flag | Primary Outcome |
| :--- | :--- | :--- |
| **Identity** | `--cluster-name` | Reliable resource tracking and cleanup in Harvester. |
| **Node Stability** | `--management-network` | Predictable network selection for cluster management. |
| **Node Stability** | `--node-ip-cidr` | Explicit promotion of specific IPs to `InternalIP`. |
| **Traffic Filtering** | `--node-exclude-ip-ranges` | Hides internal/administrative IPs from the K8s API. |
| **Service Control** | `--loadbalancer-network` | Experimental global alignment for LoadBalancer VIPs. |
| **Operational UX** | `--show-full-help-on-error` | Clean logs by suppressing lengthy help text on errors. |

---

## Enhancements: Charts

* **Support `extraArgs`**

    Allows users to pass custom arguments directly to the HCP binary for fine-tuned cluster customization.

* **Secret-based `cloud-config`**

    Adds support for sourcing `cloud-config` from Kubernetes Secrets, improving security and flexibility.

    PR: https://github.com/harvester/charts/pull/524, https://github.com/harvester/charts/pull/511

---

## Enhancements: App (Configuration Flags)

### 1. `--management-network`

* **Scenario**

    Used in multi-network environments where a VM has multiple NICs. It prevents the HCP from "guessing" which network to use for cluster management.

* **Example Input**

    `default/vlan100` or `vlan200`

* **Details**

    The value should match the network selected during guest cluster creation in the Rancher UI. If no namespace is provided, `default` is assumed. The app will fail if the format is invalid (e.g., `ns/net/extra`).

* **Fallback**

    If not set, the app chooses the first detected network.

<br>

### 2. `--node-ip-cidr`

* **Scenario**

    Crucial for multi-IP scenarios. It ensures the guest cluster runs on a predictable management/control plane network by explicitly defining which IP range should be treated as the `InternalIP`.

* **Example Input**

    `"192.168.122.0/24,2001:db8::/64"`

* **Details**

    HCP uses this guide to report the correct `InternalIP` to the Kubernetes API.

* **Fallback**

    If not set, the app picks the first valid IPv4 and IPv6 addresses as the `InternalIP`; all other addresses are marked as `ExternalIP`.

<br>

### 3. `--node-exclude-ip-ranges`

* **Scenario**

    Used to prevent specific IP ranges OR specific single IPs from being reported by the cloud provider. This is highly effective for filtering out virtual IPs, docker bridges, or static management IPs that should not be used for node communication.

* **Example Input**

    `"10.0.0.0/8,192.168.0.5,192.168.0.10"`

* **Details**

    This supports both CIDR ranges and comma-separated single IP addresses. By excluding these, you prevent internal-only or administrative IPs from being exposed to the Kubernetes API. These IPs will no longer appear in the output of `kubectl get nodes -o wide`, ensuring that only the intended routable addresses are used by the cluster components.

<br>

### 4. `--loadbalancer-network` (Experimental Global Configuration)

* **Scenario**

    Used when a guest cluster needs to expose all LoadBalancer services on a specific Harvester network (IP Pool) that is different from the management network.

* **Example Input**

    `poc/cluster-network-2-vlan300`

* **Details**

    When this flag is set, the HCP will use this specified network for **every** LoadBalancer allocation within that guest cluster.

* **Architectural Alignment**

    * **kube-vip Constraints:** This design ensures strict alignment with `kube-vip`. Since `kube-vip` typically binds to a single `vip_interface`, the HCP ensures all services are pinned to the network reachable by that interface.

    * **Safety:** It prevents "greedy" configuration errors where users might attempt to request IPs from multiple pools that the guest cluster cannot physically support.

* **Critical Implementation Note**

    * **Allocation vs. Routing:** While the Harvester LB IPAM will successfully reserve and return an IP from the requested pool, **successful traffic routing is not guaranteed by the HCP alone.**

    * **External Dependencies:** Connectivity depends entirely on the correct configuration of `kube-vip` (matching the physical interface) and external networking devices (switches/routers) to handle the target VLAN.

    * **Kernel Tuning:** In multi-NIC scenarios, users must ensure the Linux guest OS is configured to handle asymmetric routing. Specifically, kernel parameters like `net.ipv4.conf.all.rp_filter` (Reverse Path Filtering) may need to be adjusted (e.g., set to `2` or `0`) if traffic enters via one NIC and exits via another.

* **Fallback**

    If this flag is not provided, the HCP defaults to using the `--management-network`.

<br>

### 5. `--cluster-name` (Validation)

* **Scenario**

    Prevents resource identification failure and inventory collision. 

* **Details**

    For guest clusters created via Rancher, a unique and unified cluster name is mandatory for Harvester to correctly track the guest. Without a unique name, Harvester cannot reliably identify which guest cluster a request belongs to.

* **Impact**

    If the name is missing or remains the default `kubernetes`, resource **allocation and deallocation** (like LoadBalancer IPs or Volumes) may fail or cause state inconsistencies.

* **Warning Logic**

    The HCP will issue a warning during the bootstrap stage if the name is invalid. Additionally, a warning will be triggered during **every LoadBalancer allocation** to alert the user that the configuration needs to be updated to a unique value.

<br>

### 6. `--show-full-help-on-error`

* **Scenario**

    Used for debugging configuration issues. 

* **Details**

    By default, the cloud-provider framework displays a very lengthy help message whenever a flag is configured incorrectly. To keep logs clean and make troubleshooting easier, this "wall of text" is now **disabled by default**. 

* **Example Input**

    `true` (Enable this option only if you need to see the full list of available flags and descriptions during a debugging session).


## Appendix

### ccm built-in options

```
Usage:
  cloud-controller-manager [flags]

Debugging flags:

      --contention-profiling                                                                                                                                                              
                Enable block profiling, if profiling is enabled
      --profiling                                                                                                                                                                         
                Enable profiling via web interface host:port/debug/pprof/ (default true)

Leader-migration flags:

      --enable-leader-migration                                                                                                                                                           
                Whether to enable controller leader migration.
      --leader-migration-config string                                                                                                                                                    
                Path to the config file for controller leader migration, or empty to use the value that reflects default configuration of the controller manager. The config file should be
                of type LeaderMigrationConfiguration, group controllermanager.config.k8s.io, version v1alpha1.

Generic flags:

      --allocate-node-cidrs                                                                                                                                                               
                Should CIDRs for Pods be allocated and set on the cloud provider. Requires --cluster-cidr.
      --cidr-allocator-type string                                                                                                                                                        
                Type of CIDR allocator to use (default "RangeAllocator")
      --cloud-config string                                                                                                                                                               
                The path to the cloud provider configuration file. Empty string for no configuration file.
      --cloud-provider string                                                                                                                                                             
                The provider for cloud services. Empty string for no provider.
      --cluster-cidr string                                                                                                                                                               
                CIDR Range for Pods in cluster. Only used when --allocate-node-cidrs=true; if false, this option will be ignored.
      --cluster-name string                                                                                                                                                               
                The instance prefix for the cluster. (default "kubernetes")
      --configure-cloud-routes                                                                                                                                                            
                Should CIDRs allocated by allocate-node-cidrs be configured on the cloud provider. (default true)
      --controller-start-interval duration                                                                                                                                                
                Interval between starting controller managers.
      --controllers strings                                                                                                                                                               
                A list of controllers to enable. '*' enables all on-by-default controllers, 'foo' enables the controller named 'foo', '-foo' disables the controller named 'foo'.
                All controllers: cloud-node-controller, cloud-node-lifecycle-controller, node-route-controller, service-lb-controller
                Disabled-by-default controllers:  (default [*])
      --external-cloud-volume-plugin string                                                                                                                                               
                The plugin to use when cloud provider is set to external. Can be empty, should only be set when cloud-provider is external. Currently used to allow node-ipam-controller,
                persistentvolume-binder-controller, persistentvolume-expander-controller and attach-detach-controller to work for in tree cloud providers.
      --feature-gates mapStringBool                                                                                                                                                       
                A set of key=value pairs that describe feature gates for alpha/experimental features. Options are:
                APIResponseCompression=true|false (BETA - default=true)
                APIServerIdentity=true|false (BETA - default=true)
                APIServerTracing=true|false (BETA - default=true)
                APIServingWithRoutine=true|false (ALPHA - default=false)
                AllAlpha=true|false (ALPHA - default=false)
                AllBeta=true|false (BETA - default=false)
                AllowParsingUserUIDFromCertAuth=true|false (BETA - default=true)
                AllowUnsafeMalformedObjectDeletion=true|false (ALPHA - default=false)
                AnonymousAuthConfigurableEndpoints=true|false (BETA - default=true)
                AuthorizeWithSelectors=true|false (BETA - default=true)
                CBORServingAndStorage=true|false (ALPHA - default=false)
                CloudControllerManagerWebhook=true|false (ALPHA - default=false)
                ConcurrentWatchObjectDecode=true|false (BETA - default=false)
                ConsistentListFromCache=true|false (BETA - default=true)
                ContextualLogging=true|false (BETA - default=true)
                CoordinatedLeaderElection=true|false (BETA - default=false)
                ListFromCacheSnapshot=true|false (ALPHA - default=false)
                LoggingAlphaOptions=true|false (ALPHA - default=false)
                LoggingBetaOptions=true|false (BETA - default=true)
                MutatingAdmissionPolicy=true|false (ALPHA - default=false)
                OpenAPIEnums=true|false (BETA - default=true)
                RemoteRequestHeaderUID=true|false (BETA - default=true)
                ResilientWatchCacheInitialization=true|false (BETA - default=true)
                StorageVersionAPI=true|false (ALPHA - default=false)
                StorageVersionHash=true|false (BETA - default=true)
                StreamingCollectionEncodingToJSON=true|false (BETA - default=true)
                StreamingCollectionEncodingToProtobuf=true|false (BETA - default=true)
                StructuredAuthenticationConfiguration=true|false (BETA - default=true)
                UnauthenticatedHTTP2DOSMitigation=true|false (BETA - default=true)
                WatchCacheInitializationPostStartHook=true|false (BETA - default=false)
                WatchList=true|false (BETA - default=false)
      --kube-api-burst int32                                                                                                                                                              
                Burst to use while talking with kubernetes apiserver. (default 30)
      --kube-api-content-type string                                                                                                                                                      
                Content type of requests sent to apiserver. (default "application/vnd.kubernetes.protobuf")
      --kube-api-qps float32                                                                                                                                                              
                QPS to use while talking with kubernetes apiserver. (default 20)
      --leader-elect                                                                                                                                                                      
                Start a leader election client and gain leadership before executing the main loop. Enable this when running replicated components for high availability. (default true)
      --leader-elect-lease-duration duration                                                                                                                                              
                The duration that non-leader candidates will wait after observing a leadership renewal until attempting to acquire leadership of a led but unrenewed leader slot. This is
                effectively the maximum duration that a leader can be stopped before it is replaced by another candidate. This is only applicable if leader election is enabled. (default 15s)
      --leader-elect-renew-deadline duration                                                                                                                                              
                The interval between attempts by the acting master to renew a leadership slot before it stops leading. This must be less than the lease duration. This is only applicable if
                leader election is enabled. (default 10s)
      --leader-elect-resource-lock string                                                                                                                                                 
                The type of resource object that is used for locking during leader election. Supported options are 'leases'. (default "leases")
      --leader-elect-resource-name string                                                                                                                                                 
                The name of resource object that is used for locking during leader election. (default "cloud-controller-manager")
      --leader-elect-resource-namespace string                                                                                                                                            
                The namespace of resource object that is used for locking during leader election. (default "kube-system")
      --leader-elect-retry-period duration                                                                                                                                                
                The duration the clients should wait between attempting acquisition and renewal of a leadership. This is only applicable if leader election is enabled. (default 2s)
      --min-resync-period duration                                                                                                                                                        
                The resync period in reflectors will be random between MinResyncPeriod and 2*MinResyncPeriod. (default 12h0m0s)
      --node-monitor-period duration                                                                                                                                                      
                The period for syncing NodeStatus in cloud-node-lifecycle-controller. (default 5s)
      --route-reconciliation-period duration                                                                                                                                              
                The period for reconciling routes created for Nodes by cloud provider. (default 10s)
      --use-service-account-credentials                                                                                                                                                   
                If true, use individual service account credentials for each controller.

Cloud-node-controller flags:

      --concurrent-node-syncs int32                                                                                                                                                       
                Number of workers concurrently synchronizing nodes. (default 1)

Service-lb-controller flags:

      --concurrent-service-syncs int32                                                                                                                                                    
                The number of services that are allowed to sync concurrently. Larger number = more responsive service management, but more CPU (and network) load (default 1)

Webhook flags:

      --webhooks strings                                                                                                                                                                  
                A list of webhooks to enable. '*' enables all on-by-default webhooks, 'foo' enables the webhook named 'foo', '-foo' disables the webhook named 'foo'.
                All webhooks: 
                Disabled-by-default webhooks: 

Webhook serving flags:

      --webhook-bind-address ip                                                                                                                                                           
                The IP address on which to listen for the --webhook-secure-port port. The associated interface(s) must be reachable by the rest of the cluster, and by CLI/web clients. If
                set to an unspecified address (0.0.0.0 or ::), all interfaces will be used. If unset, defaults to 0.0.0.0. (default 0.0.0.0)
      --webhook-cert-dir string                                                                                                                                                           
                The directory where the TLS certs are located. If --tls-cert-file and --tls-private-key-file are provided, this flag will be ignored.
      --webhook-secure-port int                                                                                                                                                           
                Secure port to serve cloud provider webhooks. If 0, don't serve webhooks at all. (default 10260)
      --webhook-tls-cert-file string                                                                                                                                                      
                File containing the default x509 Certificate for HTTPS. (CA cert, if any, concatenated after server cert). If HTTPS serving is enabled, and --tls-cert-file and
                --tls-private-key-file are not provided, a self-signed certificate and key are generated for the public address and saved to the directory specified by --cert-dir.
      --webhook-tls-private-key-file string                                                                                                                                               
                File containing the default x509 private key matching --tls-cert-file.

Secure serving flags:

      --bind-address ip                                                                                                                                                                   
                The IP address on which to listen for the --secure-port port. The associated interface(s) must be reachable by the rest of the cluster, and by CLI/web clients. If blank or
                an unspecified address (0.0.0.0 or ::), all interfaces and IP address families will be used. (default 0.0.0.0)
      --cert-dir string                                                                                                                                                                   
                The directory where the TLS certs are located. If --tls-cert-file and --tls-private-key-file are provided, this flag will be ignored.
      --disable-http2-serving                                                                                                                                                             
                If true, HTTP2 serving will be disabled [default=false]
      --http2-max-streams-per-connection int                                                                                                                                              
                The limit that the server gives to clients for the maximum number of streams in an HTTP/2 connection. Zero means to use golang's default.
      --permit-address-sharing                                                                                                                                                            
                If true, SO_REUSEADDR will be used when binding the port. This allows binding to wildcard IPs like 0.0.0.0 and specific IPs in parallel, and it avoids waiting for the kernel
                to release sockets in TIME_WAIT state. [default=false]
      --permit-port-sharing                                                                                                                                                               
                If true, SO_REUSEPORT will be used when binding the port, which allows more than one instance to bind on the same address and port. [default=false]
      --secure-port int                                                                                                                                                                   
                The port on which to serve HTTPS with authentication and authorization. If 0, don't serve HTTPS at all. (default 10258)
      --tls-cert-file string                                                                                                                                                              
                File containing the default x509 Certificate for HTTPS. (CA cert, if any, concatenated after server cert). If HTTPS serving is enabled, and --tls-cert-file and
                --tls-private-key-file are not provided, a self-signed certificate and key are generated for the public address and saved to the directory specified by --cert-dir.
      --tls-cipher-suites strings                                                                                                                                                         
                Comma-separated list of cipher suites for the server. If omitted, the default Go cipher suites will be used. 
                Preferred values: TLS_AES_128_GCM_SHA256, TLS_AES_256_GCM_SHA384, TLS_CHACHA20_POLY1305_SHA256, TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
                TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256, TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA, TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384, TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
                TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256, TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA, TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
                TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384, TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305, TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256. 
                Insecure values: TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256, TLS_ECDHE_ECDSA_WITH_RC4_128_SHA, TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA, TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
                TLS_ECDHE_RSA_WITH_RC4_128_SHA, TLS_RSA_WITH_3DES_EDE_CBC_SHA, TLS_RSA_WITH_AES_128_CBC_SHA, TLS_RSA_WITH_AES_128_CBC_SHA256, TLS_RSA_WITH_AES_128_GCM_SHA256,
                TLS_RSA_WITH_AES_256_CBC_SHA, TLS_RSA_WITH_AES_256_GCM_SHA384, TLS_RSA_WITH_RC4_128_SHA.
      --tls-min-version string                                                                                                                                                            
                Minimum TLS version supported. Possible values: VersionTLS10, VersionTLS11, VersionTLS12, VersionTLS13
      --tls-private-key-file string                                                                                                                                                       
                File containing the default x509 private key matching --tls-cert-file.
      --tls-sni-cert-key namedCertKey                                                                                                                                                     
                A pair of x509 certificate and private key file paths, optionally suffixed with a list of domain patterns which are fully qualified domain names, possibly with prefixed
                wildcard segments. The domain patterns also allow IP addresses, but IPs should only be used if the apiserver has visibility to the IP address requested by a client. If no
                domain patterns are provided, the names of the certificate are extracted. Non-wildcard matches trump over wildcard matches, explicit domain patterns trump over extracted
                names. For multiple key/certificate pairs, use the --tls-sni-cert-key multiple times. Examples: "example.crt,example.key" or "foo.crt,foo.key:*.foo.com,foo.com". (default [])

Authentication flags:

      --authentication-kubeconfig string                                                                                                                                                  
                kubeconfig file pointing at the 'core' kubernetes server with enough rights to create tokenreviews.authentication.k8s.io. This is optional. If empty, all token requests are
                considered to be anonymous and no client CA is looked up in the cluster.
      --authentication-skip-lookup                                                                                                                                                        
                If false, the authentication-kubeconfig will be used to lookup missing authentication configuration from the cluster.
      --authentication-token-webhook-cache-ttl duration                                                                                                                                   
                The duration to cache responses from the webhook token authenticator. (default 10s)
      --authentication-tolerate-lookup-failure                                                                                                                                            
                If true, failures to look up missing authentication configuration from the cluster are not considered fatal. Note that this can result in authentication that treats all
                requests as anonymous.
      --client-ca-file string                                                                                                                                                             
                If set, any request presenting a client certificate signed by one of the authorities in the client-ca-file is authenticated with an identity corresponding to the CommonName
                of the client certificate.
      --requestheader-allowed-names strings                                                                                                                                               
                List of client certificate common names to allow to provide usernames in headers specified by --requestheader-username-headers. If empty, any client certificate validated by
                the authorities in --requestheader-client-ca-file is allowed.
      --requestheader-client-ca-file string                                                                                                                                               
                Root certificate bundle to use to verify client certificates on incoming requests before trusting usernames in headers specified by --requestheader-username-headers.
                WARNING: generally do not depend on authorization being already done for incoming requests.
      --requestheader-extra-headers-prefix strings                                                                                                                                        
                List of request header prefixes to inspect. X-Remote-Extra- is suggested. (default [x-remote-extra-])
      --requestheader-group-headers strings                                                                                                                                               
                List of request headers to inspect for groups. X-Remote-Group is suggested. (default [x-remote-group])
      --requestheader-uid-headers strings                                                                                                                                                 
                List of request headers to inspect for UIDs. X-Remote-Uid is suggested. Requires the RemoteRequestHeaderUID feature to be enabled.
      --requestheader-username-headers strings                                                                                                                                            
                List of request headers to inspect for usernames. X-Remote-User is common. (default [x-remote-user])

Authorization flags:

      --authorization-always-allow-paths strings                                                                                                                                          
                A list of HTTP paths to skip during authorization, i.e. these are authorized without contacting the 'core' kubernetes server. (default [/healthz,/readyz,/livez])
      --authorization-kubeconfig string                                                                                                                                                   
                kubeconfig file pointing at the 'core' kubernetes server with enough rights to create subjectaccessreviews.authorization.k8s.io. This is optional. If empty, all requests not
                skipped by authorization are forbidden.
      --authorization-webhook-cache-authorized-ttl duration                                                                                                                               
                The duration to cache 'authorized' responses from the webhook authorizer. (default 10s)
      --authorization-webhook-cache-unauthorized-ttl duration                                                                                                                             
                The duration to cache 'unauthorized' responses from the webhook authorizer. (default 10s)

Misc flags:

      --kubeconfig string                                                                                                                                                                 
                Path to kubeconfig file with authorization and master location information (the master location can be overridden by the master flag).
      --master string                                                                                                                                                                     
                The address of the Kubernetes API server (overrides any value in kubeconfig).
      --node-status-update-frequency duration                                                                                                                                             
                Specifies how often the controller updates nodes' status. (default 5m0s)

Global flags:

  -h, --help                                                                                                                                                                              
                help for cloud-controller-manager
      --log-flush-frequency duration                                                                                                                                                      
                Maximum number of seconds between log flushes (default 5s)
      --log-text-info-buffer-size quantity                                                                                                                                                
                [Alpha] In text format with split output streams, the info messages can be buffered for a while to increase performance. The default value of zero bytes disables buffering.
                The size can be specified as number of bytes (512), multiples of 1000 (1K), multiples of 1024 (2Ki), or powers of those (3M, 4G, 5Mi, 6Gi). Enable the LoggingAlphaOptions
                feature gate to use this.
      --log-text-split-stream                                                                                                                                                             
                [Alpha] In text format, write error messages to stderr and info messages to stdout. The default is to write a single stream to stdout. Enable the LoggingAlphaOptions feature
                gate to use this.
      --logging-format string                                                                                                                                                             
                Sets the log format. Permitted formats: "text". (default "text")
  -v, --v Level                                                                                                                                                                           
                number for the log level verbosity
      --version version[=true]                                                                                                                                                            
                --version, --version=raw prints version information and quits; --version=vX.Y.Z... sets the reported version
      --vmodule pattern=N,...                                                                                                                                                             
                comma-separated list of pattern=N settings for file-filtered logging (only works for text log format)

Harvester flags:

      --disable-annotation-alpha-provided-ip-addr                                                                                                                                         
                By default, if the 'alpha.kubernetes.io/provided-node-ip' annotation is present, the cloud-provider 
                    limits internal IP reporting to that specific address. Setting this to true causes the provider 
                    to ignore this legacy annotation and instead determine the node IP based on the discovery pipeline 
                    defined by --management-network and --node-ip-cidr.
      --disable-vmi-controller                                                                                                                                                            
                Disable sync topology to nodes and not affect the custom cluster.
      --loadbalancer-network string                                                                                                                                                       
                (Experimental) Define the Harvester network name for LoadBalancer services (e.g., 'poc/vlan300'). 
                    When set, all LoadBalancer IPs will be allocated from this specific network. Successful routing 
                    requires alignment with kube-vip configuration and potential guest OS kernel tuning.
      --management-network string                                                                                                                                                         
                Define the management network of this guest cluster, which is carried by a Harvester network 
                    (e.g., 'default/vlan-100'). This setting serves two primary purposes: 
                    1. Node IP Reporting: Guides the instance manager to the specific network interface from which 
                       to fetch the node's internal/external IP addresses. 
                    2. LoadBalancer Allocation: Guides the loadbalancer plugin to allocate Service IPs from the 
                       IPPool associated with this network.
      --node-exclude-ip-ranges strings                                                                                                                                                    
                Define IP ranges or single IPs to exclude (e.g., '10.0.0.0/8,2001:db8::/64,192.168.0.5'). This is the 
                    final safety filter; any IP matching these ranges will not be marked as InternalIP or ExternalIP. 
                    Consequently, they will be suppressed and will not appear in 'kubectl get nodes -o wide'. 
                    This global setting replaces the legacy 'cloudprovider.harvesterhci.io/additional-internal-ips' 
                    node annotation.
      --node-ip-cidr string                                                                                                                                                               
                Comma-separated list of CIDRs to filter node IPs (e.g., '192.168.122.0/24'). When used with 
                    --management-network, the instance manager will use this as a secondary filter to mark specific 
                    IPs on that interface as InternalIP. This prevents non-deterministic selection when a single 
                    network interface has multiple IP addresses.
      --show-full-help-on-error                                                                                                                                                           
                If a configuration error occurs at startup, the full help menu and flag list will be displayed. (default false)
```