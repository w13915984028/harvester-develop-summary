# How SUSE Virtualization Eliminates Virtual Machine OOM Killers

When running production workloads on virtualized infrastructure like **SUSE Virtualization**, memory management is critical. In SUSE Virtualization versions prior to v1.4.0, certain workloads experienced sudden Virtual Machine (VM) terminations due to the host Linux operating system triggering **Out-Of-Memory (OOM)** kills.

What is **KubeVirt**? SUSE Virtualization uses KubeVirt as its core virtualization engine. KubeVirt is an open-source technology that allows Kubernetes to run and manage traditional Virtual Machines inside standard containers, translating VM specifications directly into Pod configurations.

This article explores why these OOM events occur in a Kubernetes-native virtualization environment and how SUSE Virtualization provides granular tools to eliminate them.

## Anatomy of a VM OOM Event

When a VM is terminated due to insufficient memory at the host level, the Linux kernel logs specific keywords that help pinpoint the fault. In SUSE Virtualization, these logs generally fall into two distinct categories depending on which process triggered the exhaustion.

Example 1: `virt-launcher invoked oom-killer`

The `virt-launcher` process runs inside the dedicated Kubernetes Pod backing the VM. If this component or its direct sub-processes run out of the memory allocated to their cgroup, the kernel triggers a memory cgroup (memcg) OOM event.

```sh
Feb 03 19:57:08 ** kernel: virt-launcher invoked oom-killer: gfp_mask=0xcc0(GFP_KERNEL), order=0, oom_score_adj=986
Feb 03 19:57:08 ** kernel: CPU: 40 PID: 40785 Comm: virt-launcher Tainted: G          I    X    5.14.21-150400.24.60-default #1 SLE15-SP4 9096397fa6646928cc6d185ba417f2af65b536f1
...

Feb 03 19:57:08 ** kernel: memory: usage 17024340kB, limit 17024340kB, failcnt 1243
Feb 03 19:57:08 ** kernel: memory+swap: usage 17024340kB, limit 9007199254740988kB, failcnt 0                                                                 
Feb 03 19:57:08 ** kernel: kmem: usage 143556kB, limit 9007199254740988kB, failcnt 0
Feb 03 19:57:08 ** kernel: Memory cgroup stats for /kubepods.slice/kubepods-burstable.slice/kubepods-burstable-pod968a06fb_9ab9_4819_8caf_0392ddff3d9b.slice:
...
Feb 03 19:57:08 ** kernel: Tasks state (memory values in pages):
Feb 03 19:57:08 ** kernel: [  pid  ]   uid  tgid total_vm      rss pgtables_bytes swapents oom_score_adj name
Feb 03 19:57:08 ** kernel: [  38886]     0 38886      243        1    28672        0          -998 pause
Feb 03 19:57:08 ** kernel: [  38917]     0 38917   310400     6921   192512        0           986 virt-launcher-m
Feb 03 19:57:08 ** kernel: [  38934]     0 38934  1200940    25126   954368        0           986 virt-launcher
Feb 03 19:57:08 ** kernel: [  38951]     0 38951   386525     8247   466944        0           986 libvirtd
Feb 03 19:57:08 ** kernel: [  38952]     0 38952    33619     3940   290816        0           986 virtlogd
Feb 03 19:57:08 ** kernel: [  39079]   107 39079  4457263  4201766 34439168        0           986 qemu-system-x86
Feb 03 19:57:08 ** kernel: oom-kill:constraint=CONSTRAINT_MEMCG,nodemask=(null),cpuset=cri-containerd-0f32894de86edf3d3832702af794874ef8d400b4969acdea4976b12040756e0d.scope,mems_allowed=0-1,oom_memcg=/kubepods.slice/kubepods-burstable.slice/kubepods-burstable-pod968a06fb_9ab9_4819_8caf_0392ddff3d9b.slice,task_memcg=/kubepods.slice/kubepods-burstable.slice/kubepods-burstable-pod968a06fb_9ab9_4819_8caf_0392ddff3d9b.slice/cri-containerd-0f32894de86edf3d3832702af794874ef8d400b4969acdea4976b12040756e0d.scope,task=qemu-system-x86,pid=39079,uid=107
```

Example 2: `CPU X/KVM invoked oom-killer`

This occurs when a vCPU execution thread inside qemu-system-x86_64 attempts a memory operation that pushes the entire container beyond its Kubernetes memory limit.

```sh
[Thu May  9 14:52:38 2024] CPU 11/KVM invoked oom-killer: gfp_mask=0xcc0(GFP_KERNEL), order=0, oom_score_adj=830
[Thu May  9 14:52:38 2024] CPU: 60 PID: 70888 Comm: CPU 11/KVM Not tainted 5.3.18-150300.59.101-default #1 SLE15-SP3
...
[Thu May  9 14:52:38 2024] memory: usage 67579904kB, limit 67579904kB, failcnt 67391
[Thu May  9 14:52:38 2024] memory+swap: usage 0kB, limit 9007199254740988kB, failcnt 0
[Thu May  9 14:52:38 2024] kmem: usage 633636kB, limit 9007199254740988kB, failcnt 0
...
[Thu May  9 14:52:38 2024] Tasks state (memory values in pages):
[Thu May  9 14:52:38 2024] [  pid  ]   uid  tgid total_vm      rss pgtables_bytes swapents oom_score_adj name
[Thu May  9 14:52:38 2024] [  70675]     0 70675      243        1    28672        0          -998 pause
[Thu May  9 14:52:38 2024] [  70728]     0 70728   310400     5467   188416        0           830 virt-launcher-m
[Thu May  9 14:52:38 2024] [  70746]     0 70746  1242373    25104  1073152        0           830 virt-launcher
[Thu May  9 14:52:38 2024] [  70762]     0 70762   455279    14110   770048        0           830 libvirtd
[Thu May  9 14:52:38 2024] [  70763]     0 70763    37704     3916   339968        0           830 virtlogd
[Thu May  9 14:52:38 2024] [  70870]   107 70870 18302464 16718510 135278592        0           830 qemu-system-x86
[Thu May  9 14:52:38 2024] oom-kill:constraint=CONSTRAINT_MEMCG,nodemask=(null),cpuset=cri-containerd-100093783c22a3ae1a42e21dd887b7c26eef52d56ba44c7273ef54507b6efe7c.scope,mems_allowed=0-3,oom_memcg=/kubepods.slice/kubepods-burstable.slice/kubepods-burstable-podef91e487_dec5_4613_800b_eb23e1a1617d.slice,task_memcg=/kubepods.slice/kubepods-burstable.slice/kubepods-burstable-podef91e487_dec5_4613_800b_eb23e1a1617d.slice/cri-containerd-100093783c22a3ae1a42e21dd887b7c26eef52d56ba44c7273ef54507b6efe7c.scope,task=qemu-system-x86,pid=70870,uid=107
[Thu May  9 14:52:38 2024] Memory cgroup out of memory: Killed process 70870 (qemu-system-x86) total-vm:73209856kB, anon-rss:66852088kB, file-rss:21948kB, shmem-rss:4kB
[Thu May  9 14:52:38 2024] oom_reaper: reaped process 70870 (qemu-system-x86), now anon-rss:0kB, file-rss:132kB, shmem-rss:4kB
```

## Root Cause: Native Hypervisors vs. SUSE Virtualization Architecture

### Traditional Linux Host (e.g., Virtual Machine Manager)

On a standard Linux host, a VM managed via QEMU/KVM runs inside a systemd machine.slice. The hypervisor process (qemu-system-x86_64) has access to the host's wider pool of resources, managed loosely unless strict cgroup limits are manually added.

```sh
$ systemd-cgls
Control group /:
-.slice
├─1173 bpfilter_umh
├─system.slice

└─machine.slice
  └─machine-qemu\x2d1\x2dharv41.scope 
    └─8632 /usr/bin/qemu-system-x86_64 -name guest=harv41,debug-threads=on -S -…
```

### SUSE Virtualization (Kubernetes/KubeVirt Engine)

In SUSE Virtualization, every VM is encapsulated inside a Kubernetes Pod. This introduces a strict cgroup boundary (kubepods.slice).

As shown below, multiple helper processes must live alongside the primary qemu-system-x86_64 process within the same tightly limited container memory budget:

```sh
-.slice
└─kubepods.slice
  │ ├─kubepods-burstable-pod99ee3a64_645b_4699_9384_5a3875d78b41.slice
  │ │ ├─cri-containerd-0c316eb8a4711bff1ce968b46ddb658e49378897454b4caf2a20704c808f33f1.scope …
  │ │ │ └─ 8505 /pause
  │ │ ├─cri-containerd-eb5deef29f064adfd2456d9f9c535674ac4d0c95c81b13afbdab5a89dc6a774b.scope …
  │ │ │ └─ 8590 /usr/bin/virt-tail --logfile /var/run/kubevirt-private/2ce151aa…
  │ │ └─cri-containerd-fd57a5cfc2b9b1f53eaf7b575c3273e6784f4c56a04a17d502ecfbd19e55b066.scope …
  │ │   ├─ 8542 /usr/bin/virt-launcher-monitor --qemu-timeout 301s --name vm2 -…
  │ │   ├─ 8558 /usr/bin/virt-launcher --qemu-timeout 301s --name vm2 --uid 2ce…
  │ │   ├─ 8591 /usr/sbin/virtqemud -f /var/run/libvirt/virtqemud.conf
  │ │   ├─ 8592 /usr/sbin/virtlogd -f /etc/libvirt/virtlogd.conf
  │ │   └─ 8823 /usr/bin/qemu-system-x86_64 -name guest=default_vm2,debug-threa…
```

### The "Hidden" Memory Overhead

#### Breaking Down the Memory Overhead Buffer

When you define a virtual machine—for example, a VM configured with **4 vCPUs, 2 GiB of memory, and 1 Ethernet interface**—KubeVirt does not just allocate exactly 2 GiB of memory to the container. 

Instead, KubeVirt calculates an additional baseline memory overhead required to operate the virtualization stack. This overhead budget covers:

*   **CPU Simulators:** Thread pools tracking guest state and handling context switches.
*   **Memory Management:** Tracking structures such as QEMU page tables mapping guest RAM.
*   **Auxiliary Devices:** Buffers for virtual network interfaces (NICs), storage queues, and video devices.

#### The High Stakes of VM OOM Kills

Depending on the guest OS type, specific kernel workloads, and heavy storage/network I/O spikes, the memory consumed by these helper tasks can quickly exceed KubeVirt's default calculations. Because Kubernetes enforces a strict hard ceiling on the Pod container, the entire container triggers a `CONSTRAINT_MEMCG` OOM kill the moment this boundary is breached.

Unlike traditional, stateless Kubernetes workloads where a container crash is quickly mitigated by a rapid pod restart, an OOM kill on a VM pod carries severe operational consequences:

* **Prolonged Downtime:** A virtual machine is a stateful workload. It does not instantaneously serve traffic upon a container restart; it must undergo a full operating system boot cycle, run init scripts, and re-initialize services, drastically extending your Recovery Time Objective (RTO).

* **Risk of Data Corruption:** Sudden terminations during flight can abruptly cut off active storage queues. If the guest OS or database engine is in the middle of a critical write operation when the host terminates the `qemu` process, it can result in uncommitted journals, filesystem degradation, or severe data corruption on your persistent volumes.

## The Solution: Tunable Memory Architectures

To address this, SUSE Virtualization introduced dual-layer configurations that give administrators full flexibility over how overhead buffers are calculated.

### Global Adjustment: additional-guest-memory-overhead-ratio

This cluster-wide setting functions as a multiplier for KubeVirt's automatically calculated memory overhead. For deep structural details, refer to the [SUSE Virtualization Advanced Documentation](https://docs.harvesterhci.io/v1.8/advanced/index#additional-guest-memory-overhead-ratio).

*   **Definition:** Scales the calculated overhead buffer to accommodate heavy I/O or virtualization tasks.
*   **Default Value:** `1.5` (Provides a 50% safety cushion above baseline calculations).
*   **Valid Range:** `0` or `1.0` to `10.0`.

> 💡 **Important Operational Notes:**
> 
> *   **Lifecycle Impact:** Changes to this setting only apply to newly created virtual machines or existing VMs after they undergo a migration or a full power cycle.
> *   **System Overhead:** A higher ratio increases the host container's memory allocation, guaranteeing safety for heavy workloads but scaling up the overall system resource reservation footprint.
> *   **Resource Allocation Trade-off:** Setting this ratio excessively high can lock up unneeded host memory blocks, leading to predictable underutilization and significant memory waste across your compute nodes.

### Per-VM Optimization: Reserved Memory

For specific virtual machines running intensive or non-standard workloads, a global multiplier might not offer the precision required. SUSE Virtualization allows administrators to define a dedicated `Reserved Memory` value directly on individual VMs. For complete configuration steps, see the [SUSE Virtualization VM Management Documentation](https://docs.harvesterhci.io/v1.8/vm/index#reserved-memory).

> ⚠️ **Under the Hood Memory Carving:**
> When you configure this setting, SUSE Virtualization explicitly scales down the available memory presented to the Guest OS inside the VM. For example, if a VM is configured with **2 GiB** of memory and you set a Reserved Memory value of **256 MiB**, the Guest OS will only see and utilize **1.75 GiB** (`2 GiB - 256 MiB`).

This mechanism implicitly and securely saves that carved-out space exclusively for helper tasks, altering the memory calculation logic as follows:

**Total Memory Overhead** = **Auto-calculated Overhead**  * **Ratio** + **Reserved Memory**

#### Why Use Per-VM Reserved Memory?

* **Guaranteed Overhead Headroom:** By restricting the Guest OS from consuming the top slice of its configured allocation, you guarantee an isolated, un-evictable memory runway for host helper tasks.

* **Targeted Safety for Heavy Workloads:** This mechanism is highly practical for mission-critical, high-performance, or special-purpose workloads (such as nested virtualization layers or intensive database engines). It effectively prevents the VM from running into host-level cgroup OOM termination by proactively limiting its internal usage boundaries, removing the risk of unexpected node-level kills.

* **Optimized Cluster Usability:** Using per-VM reservations eliminates the major disadvantage of cranking up the global `additional-guest-memory-overhead-ratio` for the whole cluster. Instead of forcing a massive, wasteful memory overhead reservation across *every* idle or lightweight VM on your hosts, you can maintain a lean global default and surgically protect only the heavy workloads—striking an ideal balance between system density and ironclad stability.

## Best Practices & Configuration Matrix

The following matrix showcases how combinations of **Reserved Memory** and the **Overhead Ratio** change the actual layout of the Guest OS space versus what Kubernetes reserves as a hard boundary.

| VM Configured Memory | Reserved Memory | additional-guest-memory-overhead-ratio| Guest OS Memory | POD Container Memory Limit | Total Memory Overhead |
| --- | --- | --- | --- | --- | --- |
| 2 Gi | not configured | "0.0" | 2 Gi - 100 Mi | 2 Gi + 240 Mi | ~340 Mi |
| 2 Gi | 256 Mi | "0.0" | 2 Gi - 256 Mi | 2 Gi + 240 Mi | ~500 Mi |
| 2 Gi | not configured | "1.0" | 2 Gi | 2 Gi + 240*1.0 Mi | ~240 Mi |
| 2 Gi | not configured | "3.0" | 2 Gi | 2 Gi + 240*3.0 Mi | ~720 Mi |
| 2 Gi | not configured | "1.5" | 2 Gi | 2 Gi + 240*1.5 Mi | ~360 Mi |
| 2 Gi | 256 Mi | "1.5" | 2 Gi - 256 Mi | 2 Gi + 240*1.5 Mi | ~620 Mi |

When optimizing your SUSE Virtualization cluster to eliminate host-level container OOM events, use the following operational checklist to tailor your memory strategies:

*   **For General Workloads:**
    *   Stick to the default ratio of `1.5`, or configure a slightly higher value of `2.0`. This ensures that standard Guest operating systems receive exactly the memory requested while scaling out a stable, predictable background overhead buffer across the cluster.

*   **For High I/O and Storage-Heavy VMs:**
    *   If you observe periodic KVM OOM events during massive backup windows, large-scale data syncs, or intensive disk read/write cycles, increase the individual VM's allocation or implement a targeted `Reserved Memory` configuration to safely expand the helper overhead pool.

*   **For GPU Passthrough Workloads:**
    *   Virtual machines utilizing direct hardware acceleration or GPU passthrough are prime candidates for explicit `Reserved Memory` carving. The underlying host-side device drivers and memory-mapped I/O (MMIO) windows for high-performance graphics hardware require a significantly higher, specialized memory footprint outside the guest OS space. Allocating dedicated per-VM reserved memory prevents driver-instigated cgroup allocation breaches, keeping both the hardware pipeline and the hypervisor completely stable.

## Quick Summary

*   **The Problem:** In SUSE Virtualization's Kubernetes-native architecture, every virtual machine is bound by a strict Pod container limit. While this rigid cgroup boundary is essential for security—ensuring a single rogue or leaking VM can never starve neighboring workloads or crash the bare-metal host—it means heavy storage/network I/O, device drivers, or GPU passthrough can cause internal helper processes to breach this hard ceiling, triggering a sudden host-level OOM kill.

*   **The Solution:** SUSE Virtualization eliminates these crashes without losing secure resource control using a dual-layer memory tuning strategy:
    *   **Globally:** The `additional-guest-memory-overhead-ratio` scales out a safety cushion cluster-wide for newly created or migrated VMs.
    *   **Per-VM:** The `Reserved Memory` setting surgically carves out a chunk of the VM's configured RAM exclusively for background helper tasks—preventing wasteful memory reservations across the cluster while safely anchoring high-performance, mission-critical workloads.

## Appendix: Lab Simulation — Manually Triggering the Host-Level OOM

For engineers looking to validate this behavior safely in a staging environment, you can replicate this multi-process cgroup breach. A detailed script and case study can be found in the [Harvester Development Summary: OOM Investigation](https://github.com/w13915984028/harvester-develop-summary/blob/main/oom-related-issue-investigation.md#hold-memory-on-vm-pod-and-exhaust-memory-on-guest-vm).

The simulation process highlights a fundamental truth about modern virtualization boundaries:

*   **The Guest OS is Trustworthy:** Testing shows that modern guest operating systems handle internal resource limits reliably. If a runaway application inside the guest OS eats up all available RAM, the guest kernel safely steps in and kills that specific process internally. The VM itself survives, and from the host's perspective, the virtual machine continues running normally.

*   **The Host Cgroup Boundary is the Weak Link:** The true host-level crash only happens if processes inside the host cgroup expand unexpectedly. If an infrastructure task or helper process inside the Pod container balloons, it consumes the memory buffer that KubeVirt set aside, causing the entire cgroup—the VM's carrier—to slam into the hard Kubernetes ceiling and trigger a host-level OOM kill.
