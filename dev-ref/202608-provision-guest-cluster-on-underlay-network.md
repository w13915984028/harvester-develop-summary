# Harvester Overlay Network Guest Cluster Provisioning: Open Points & Architectural Gaps

This document details the architectural requirements, observed gaps, and proposed solutions for provisioning RKE2 guest clusters on **Harvester Overlay Networks (Kube-OVN)** while maintaining complete network isolation and self-service external reachability.

---

## 0. Core Capabilities Provided by Harvester

* **Virtual Machine & Infrastructure Provisioning:** Automated lifecycle management and orchestration of virtualized node resources.
* **Node Metadata Synchronization:** [Real-time reporting](https://github.com/harvester/docs/pull/1078) and registration of node hostnames, IP addresses, and topology labels.
* **Integrated Load Balancing:** Dynamic provision and routing of external/internal network traffic for `Type: LoadBalancer` services.
* **Cloud Storage Integration (CSI Driver):** Automated volume provisioning, dynamic attachment, and persistent storage management for guest workloads.

---

## 1. Network Topology Reference

The baseline environment consists of a physical/underlay management network (`192.168.100.0/24`) and a virtualized overlay network (`172.0.100.0/24`) hosted on Harvester:

| Component | Network Boundary | IP Address / Range | Function / Description |
| :--- | :--- | :--- | :--- |
| **Rancher Manager** | Underlay | `192.168.100.10` | Central management plane for downstream guest clusters. |
| **Harvester VIP** | Underlay | `192.168.100.20` | Harvester management cluster control plane endpoint. |
| **VPC NAT Gateway** | Boundary | `192.168.100.100` | Kube-OVN VPC Gateway with **both SNAT and DNAT enabled**. |
| **External EIP Pool** | Underlay | `192.168.100.200–254` | Managed underlay IP pool reserved for external mapping. |
| **Guest Cluster VMs** | Overlay | `172.0.100.0/25` | Virtual machines acting as RKE2 control plane / worker nodes. |
| **Guest LB Services** | Overlay | `172.0.100.0/25` | Kubernetes `Type: LoadBalancer` cluster internal VIP allocations. |
| **RWX Storage (Whereabouts)** | Overlay / Underlay | `172.0.100.128/25` | Longhorn NFS export endpoints on Harvester host storage layer. |

---

## 2. Industry Standard Pattern: Cloud Gateway EIP vs. Underlay Routing

Major public cloud providers (AWS VPC, GCP VPC) never expose internal overlay subnets to physical network core switches. Manually programming upstream routes on physical routers for private overlay subnets causes severe operational bottlenecks:
* It causes routing table pollution and management complexity on core switches.
* It breaks multi-tenancy by preventing separate clusters from reusing duplicate IP space (`172.0.100.0/24`).

**Architectural Principle:** Following public cloud architecture (AWS IGW/NAT EIP pattern), Harvester overlay networks must treat internal subnets (`172.0.100.0/24`) as completely hidden isolated spaces. External reachability must be achieved **exclusively via underlay Elastic IP (EIP) allocation and dynamic Kube-OVN 1:1 DNAT translation** at the VPC Gateway.

---

## 3. Gap Analysis & Open Points

### Open Point 1: Outbound SNAT vs. Inbound Direct Node Access (SSH / Debugging)

#### Current State & Problem
* During guest cluster provisioning, Harvester VMs boot up inside the overlay subnet (`172.0.100.*`) and successfully establish outbound connectivity to Rancher Manager (`192.168.100.10`) via the VPC NAT Gateway (`192.168.100.100`) using SNAT.
* However, when RKE2 setup scripts hang or fail, operators need to initiate inbound `SSH` connections directly to the guest VMs from external networks.
* **With DNAT Enabled:** Enabling DNAT at the gateway provides the underlying translation capability, but inbound connections still fail out of the box because **dynamic DNAT rules are not programmed automatically** during VM provisioning. Traffic destined for an unmapped `172.0.100.*` IP remains unreachable from the outside.

> **Troubleshooting Workaround Note:** For manual administrative debugging, operators can manually add static 1:1 DNAT rules (or port-forwarding mappings) on Harvester / Kube-OVN to temporarily expose an overlay node endpoint (`172.0.100.X:22`) on an underlay `192.168.100.Y` address.

#### Required Resolution
* **Automated Node EIP/DNAT Rule Orchestration:** Harvester / Rancher Node Driver must support automatically triggering dynamic Kube-OVN DNAT rule creation during VM provisioning to expose SSH/management endpoints (`172.0.100.X:22`) on an allocated EIP or gateway port.

---

### Open Point 2: Automated LoadBalancer Support (Underlay EIP Mapping via Dynamic DNAT)

#### Current State & Problem
* When a user creates a Kubernetes service with `spec.type: LoadBalancer` inside the guest cluster, the in-cluster load balancer controller (e.g., Kube-VIP / MetalLB) assigns an internal IP from the overlay range (`172.0.100.*`).
* By default, `spec.type: LoadBalancer` assumes external, direct reachability. Even though the VPC NAT Gateway has DNAT capability enabled, **dynamic DNAT rules are still needed** to link the overlay service IP with an underlay EIP. Without an automated controller loop, external clients cannot connect to the service.

#### Open Technical Question: Does Default Kube-VIP (ARP Broadcast) Work in Kube-OVN Overlay Subnets?
**Short Answer: No, but due to internal OVN port security and lack of boundary translation—not because GARP fails to cross the router.**

While `kube-vip` relies on **Gratuitous ARP (GARP)** to float Virtual IPs across a Layer-2 Ethernet segment, attempting to use standard `kube-vip` in a Kube-OVN VPC overlay environment fails due to the following mechanics:

1. **Strict IPAM & Port Security (Allowed Address Pairs):**
   Kube-OVN enforces strict **Port Security** on virtual machine interfaces. Each VM's logical port is explicitly locked to its assigned IP/MAC pair. When `kube-vip` floats an additional VIP (`172.0.100.50`) to a node's virtual interface (`eth0`), Kube-OVN drops egress traffic originating from that unallocated VIP as a port security violation (unless `allowed_address_pairs` or OVN VIP custom resources are explicitly declared).
2. **Overlay L2 Broadcast Suppression:**
   Kube-OVN logical switches suppress native L2 broadcast/multicast traffic using OpenFlow rules at the hypervisor level. While GARP updates local overlay neighbor tables, OVS may intercept or throttle flooded GARP frames unless explicit VIP resources are registered in the OVN control plane.
3. **L3 NAT Reachability Mismatch (The `kube-vip` Architectural Limit):**
   GARP operates strictly within the local L2 overlay segment to inform neighbor VMs and the VPC Gateway of the VIP owner. However, external clients on the underlay (`192.168.100.0/24`) cannot route directly to `172.0.100.50`. `kube-vip` alone cannot negotiate or request an underlay EIP (`192.168.100.220`) from Harvester, leaving the service unexposed externally despite successful internal ARP resolution.

> **Operational Note on `kube-vip` BGP Mode:**
> Switching `kube-vip` to **BGP Mode** (peering with the Kube-OVN VPC Logical Router at `172.0.100.1`) bypasses L2 ARP suppression and handles internal L3 failover cleanly without port security drops. However, **BGP mode is not plug-and-play**; it introduces operational overhead requiring explicit ASN planning, BGP peer configuration on both `kube-vip` and Kube-OVN, and dynamic route filtering. Furthermore, it only resolves *internal* overlay reachability—external underlay exposure still mandates automated EIP allocation and Gateway DNAT rules via the Harvester Cloud Controller Manager (CCM).


```text
+---------------------------------------------------------------------------------+
| 1. User Creates Service (Type: LoadBalancer)                                   |
|    - Internal Guest Service IP Allocated: 172.0.100.50                          |
+---------------------------------------------------------------------------------+
                                       |
                                       v
+---------------------------------------------------------------------------------+
| 2. Harvester Cloud Provider (CCM) Intercepts Request                           |
|    - Guest CCM requests an underlay EIP from Harvester API IPAM Pool           |
+---------------------------------------------------------------------------------+
                                       |
                                       v
+---------------------------------------------------------------------------------+
| 3. Dynamic DNAT Rule Provisioned on VPC NAT Gateway                             |
|    - Kube-OVN programs stateful translation:                                   |
|      [ EIP: 192.168.100.220:<Port> ]  ===( DNAT )===>  [ Overlay: 172.0.100.50:<Port> ] |
+---------------------------------------------------------------------------------+
                                       |
                                       v
+---------------------------------------------------------------------------------+
| 4. Service Ingress Status Updated                                               |
|    - Service status sets: status.loadBalancer.ingress[0].ip = 192.168.100.220   |
+---------------------------------------------------------------------------------+
```


#### Required Infrastructure Capabilities
To satisfy user expectations for overlay `Type: LoadBalancer` services without breaking overlay network boundaries, the following automated components are required:

1. **Harvester Underlay EIP Pool Management:** An underlay IPAM module within Harvester to allocate and track physical network VIPs (`192.168.100.200–254`).
2. **Harvester Cloud Provider (CCM) Integration:** The guest cluster's Cloud Controller Manager must intercept `Type: LoadBalancer` services and request an underlay EIP from the parent Harvester API.
3. **Automated Dynamic DNAT Orchestration:** Automatic translation rule creation at the VPC Gateway linking `EIP (192.168.100.X)` to `Internal LB IP (172.0.100.Y)` via Kube-OVN custom resources.


### Open Point 3: CSI Driver & RWX Volume Storage Support over Overlay

#### Current State & Problem
Harvester ReadWriteMany (RWX) volumes rely on Longhorn NFS target pods exported on the host cluster (using `rwx-network` or `share-storage-network`). Connecting guest workloads on overlay networks to host storage introduces two major system-level architectural challenges:

#### Challenge A: Coexistence Conflict (Overlay vs. Underlay Workloads)
When underlay (VLAN) and overlay (VPC) workloads coexist on the same Harvester host cluster:
* **Storage on Underlay VLAN:** Underlay VMs mount storage natively over L2. Overlay VMs must cross the VPC NAT Gateway via SNAT, breaking stateful NFS v3/v4 lock daemons (`NLM`/`lockd`) and causing frequent `Stale File Handle` drops.
* **Storage on Overlay:** Overlay VMs connect natively, but host nodes and underlay VMs cannot reach storage without hairpinning through tenant VPC routers, breaking multi-tenant isolation boundaries.

#### Challenge B: Pure Overlay Mode with Non-Overlapping Subnets (e.g., `/25` Split)
Even if all VMs run on pure overlay mode and subnets are strictly separated (e.g., Guest DHCP on `172.0.100.0/25` and Whereabouts Storage IPAM on `172.0.100.128/25`):
1. **Routing Boundary Overhead:** Storage traffic between the two `/25` subnets still traverses the Kube-OVN logical router. Encapsulation overhead reduces effective MTU and degrades I/O performance unless Jumbo Frames are end-to-end configured.
2. **Failover Connection Freezes:** If an OVN logical router switches active chassis nodes during host failover, stateful NFS TCP sessions freeze, causing guest Linux pods to hang indefinitely.
3. **Multi-Tenant Storage Leakage:** Placing storage targets into a common overlay network breaks tenant VPC boundaries, allowing separate overlay tenants to reach each other's storage export endpoints.

```text
+---------------------------------------------------------------------------------+
| Dual-Mode Storage Binding Dilemma                                              |
+---------------------------------------------------------------------------------+
| Option A: Bind Storage Network to Underlay VLAN                                 |
|   - Underlay Workloads:  [ OK ] Direct L2 Connectivity                          |
|   - Overlay Guest VMs:   [ FAIL ] Breaks over VPC NAT Gateway / SNAT            |
+---------------------------------------------------------------------------------+
| Option B: Bind Storage Network to Tenant Overlay VPC                            |
|   - Overlay Guest VMs:   [ OK ] Direct Logical Router Path                      |
|   - Underlay / Host:     [ FAIL ] Cannot cross tenant boundary without hairpin  |
+---------------------------------------------------------------------------------+
```

#### Required Infrastructure Capabilities
1. **Direct Storage Network Bypass (Multus Integration):** Guest VMs requiring high-performance RWX storage should utilize a secondary Multus network interface connected directly to the underlay `rwx-network`, bypassing the overlay VPC NAT Gateway entirely.
2. **Harvester CSI Driver Overlay-Aware Mounting:** The `harvester-csi-driver` running inside the guest cluster must be updated to handle storage target resolution across the overlay boundary, ensuring storage endpoints map cleanly to host-level Longhorn volumes.

---

## 4. Summary of Core Requirements

| Domain | Underlay (L2 VLAN) | Overlay (Kube-OVN VPC) | Observed Gap | Required Capability / Resolution |
| :--- | :--- | :--- | :--- | :--- |
| **Basic VM Provisioning** | Supported | Supported | None (Fully supported across both network modes). | Native node driver and cloud-init provisioning function out of the box in both network modes. |
| **Node Provisioning Access** | Direct Access | Dynamic DNAT Required | Cannot initiate inbound `SSH` to overlay nodes without dynamic DNAT rules. | Automate dynamic EIP/DNAT mapping per node during provisioning. <br>*(Workaround: Manually add static DNAT rules on Harvester).* |
| **LoadBalancer External IP** | Direct VIP IPAM | Dynamic EIP + DNAT Required | Service status receives non-routable `172.0.100.*` IP unless dynamic DNAT rule is created. | Harvester CCM must request an underlay `192.168.100.*` EIP for guest LB services. |
| **In-Cluster Load Balancer (kube-vip)** | Native L2 GARP | BGP Mode / VIP CRD Required | L2 ARP mode fails due to OVS broadcast suppression, strict port security, and lack of external L3 reachability. | Shift to BGP mode peered with VPC router for internal VIPs, coupled with Harvester CCM for external EIP/DNAT mapping. |
| **Gateway Orchestration** | N/A (L2 Direct) | Dynamic Controller Loop | Enabling DNAT on the gateway is static; runtime dynamic rule generation does not exist for guest objects. | Automate Kube-OVN dynamic DNAT rule programming when EIPs are assigned to overlay services or nodes. |
| **CSI / RWX Storage Support** | Native L2 Mount | Gateway Bottleneck / Drops | Dual-mode workloads conflict over storage binding; pure overlay `/25` splits cause stateful NFS drops and isolation leaks. | Integrate Multus secondary NICs for underlay storage traffic bypass OR update Harvester CSI driver for overlay awareness. |
