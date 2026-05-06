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
