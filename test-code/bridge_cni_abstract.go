


type CNI interface {
	AddNetworkList(ctx context.Context, net *NetworkConfigList, rt *RuntimeConf) (types.Result, error)
	CheckNetworkList(ctx context.Context, net *NetworkConfigList, rt *RuntimeConf) error
	DelNetworkList(ctx context.Context, net *NetworkConfigList, rt *RuntimeConf) error
	GetNetworkListCachedResult(net *NetworkConfigList, rt *RuntimeConf) (types.Result, error)
	GetNetworkListCachedConfig(net *NetworkConfigList, rt *RuntimeConf) ([]byte, *RuntimeConf, error)

	AddNetwork(ctx context.Context, net *PluginConfig, rt *RuntimeConf) (types.Result, error)
	CheckNetwork(ctx context.Context, net *PluginConfig, rt *RuntimeConf) error
	DelNetwork(ctx context.Context, net *PluginConfig, rt *RuntimeConf) error
	GetNetworkCachedResult(net *PluginConfig, rt *RuntimeConf) (types.Result, error)
	GetNetworkCachedConfig(net *PluginConfig, rt *RuntimeConf) ([]byte, *RuntimeConf, error)

	ValidateNetworkList(ctx context.Context, net *NetworkConfigList) ([]string, error)
	ValidateNetwork(ctx context.Context, net *PluginConfig) ([]string, error)

	GCNetworkList(ctx context.Context, net *NetworkConfigList, args *GCArgs) error
	GetStatusNetworkList(ctx context.Context, net *NetworkConfigList) error

	GetCachedAttachments(containerID string) ([]*NetworkAttachment, error)

	GetVersionInfo(ctx context.Context, pluginType string) (version.PluginInfo, error)
}







https://github.com/containernetworking/plugins/blob/main/plugins/main/bridge/bridge.go




type NetConf struct {
	types.NetConf
	BrName                    string       `json:"bridge"`
	IsGW                      bool         `json:"isGateway"`
	IsDefaultGW               bool         `json:"isDefaultGateway"`
	ForceAddress              bool         `json:"forceAddress"`
	IPMasq                    bool         `json:"ipMasq"`
	MTU                       int          `json:"mtu"`
	HairpinMode               bool         `json:"hairpinMode"`
	PromiscMode               bool         `json:"promiscMode"`
	Vlan                      int          `json:"vlan"`
	VlanTrunk                 []*VlanTrunk `json:"vlanTrunk,omitempty"`
	PreserveDefaultVlan       bool         `json:"preserveDefaultVlan"`
	MacSpoofChk               bool         `json:"macspoofchk,omitempty"`
	EnableDad                 bool         `json:"enabledad,omitempty"`
	DisableContainerInterface bool         `json:"disableContainerInterface,omitempty"`

	Args struct {
		Cni BridgeArgs `json:"cni,omitempty"`
	} `json:"args,omitempty"`
	RuntimeConfig struct {
		Mac string `json:"mac,omitempty"`
	} `json:"runtimeConfig,omitempty"`

	mac   string
	vlans []int
}


func setupBridge(n *NetConf) (*netlink.Bridge, *current.Interface, error) {
	vlanFiltering := false
	if n.Vlan != 0 || n.VlanTrunk != nil {
		vlanFiltering = true
	}
	// create bridge if necessary
	br, err := ensureBridge(n.BrName, n.MTU, n.PromiscMode, vlanFiltering)




func main() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, bv.BuildString("bridge"))
}







func cmdAdd(args *skel.CmdArgs) error {
	success := false

	n, cniVersion, err := loadNetConf(args.StdinData, args.Args)
	if err != nil {
		return err
	}

	isLayer3 := n.IPAM.Type != ""

	if isLayer3 && n.DisableContainerInterface {
		return fmt.Errorf("cannot use IPAM when DisableContainerInterface flag is set")
	}

	if n.IsDefaultGW {
		n.IsGW = true
	}

	if n.HairpinMode && n.PromiscMode {
		return fmt.Errorf("cannot set hairpin mode and promiscuous mode at the same time")
	}

	br, brInterface, err := setupBridge(n)
	if err != nil {
		return err
	}

	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer netns.Close()

	hostInterface, containerInterface, err := setupVeth(netns, br, args.IfName, n.MTU, n.HairpinMode, n.Vlan, n.vlans, n.PreserveDefaultVlan, n.mac)
	if err != nil {
		return err
	}

	// Assume L2 interface only
	result := &current.Result{
		CNIVersion: current.ImplementedSpecVersion,
		Interfaces: []*current.Interface{
			brInterface,
			hostInterface,
			containerInterface,
		},
	}

	if n.MacSpoofChk {
		sc := link.NewSpoofChecker(hostInterface.Name, containerInterface.Mac, uniqueID(args.ContainerID, args.IfName))
		if err := sc.Setup(); err != nil {
			return err
		}
		defer func() {
			if !success {
				if err := sc.Teardown(); err != nil {
					fmt.Fprintf(os.Stderr, "%v", err)
				}
			}
		}()
	}

	if isLayer3 {
		// run the IPAM plugin and get back the config to apply
		r, err := ipam.ExecAdd(n.IPAM.Type, args.StdinData)
		if err != nil {
			return err
		}

		// release IP in case of failure
		defer func() {
			if !success {
				ipam.ExecDel(n.IPAM.Type, args.StdinData)
			}
		}()

		// Convert whatever the IPAM result was into the current Result type
		ipamResult, err := current.NewResultFromResult(r)
		if err != nil {
			return err
		}

		result.IPs = ipamResult.IPs
		result.Routes = ipamResult.Routes
		result.DNS = ipamResult.DNS

		if len(result.IPs) == 0 {
			return errors.New("IPAM plugin returned missing IP config")
		}

		// Gather gateway information for each IP family
		gwsV4, gwsV6, err := calcGateways(result, n)
		if err != nil {
			return err
		}

		// Configure the container hardware address and IP address(es)
		if err := netns.Do(func(_ ns.NetNS) error {
			if n.EnableDad {
				_, _ = sysctl.Sysctl(fmt.Sprintf("/net/ipv6/conf/%s/enhanced_dad", args.IfName), "1")
				_, _ = sysctl.Sysctl(fmt.Sprintf("net/ipv6/conf/%s/accept_dad", args.IfName), "1")
			} else {
				_, _ = sysctl.Sysctl(fmt.Sprintf("net/ipv6/conf/%s/accept_dad", args.IfName), "0")
			}
			_, _ = sysctl.Sysctl(fmt.Sprintf("net/ipv4/conf/%s/arp_notify", args.IfName), "1")

			// Add the IP to the interface
			return ipam.ConfigureIface(args.IfName, result)
		}); err != nil {
			return err
		}

		if n.IsGW {
			var vlanInterface *current.Interface
			// Set the IP address(es) on the bridge and enable forwarding
			for _, gws := range []*gwInfo{gwsV4, gwsV6} {
				for _, gw := range gws.gws {
					if n.Vlan != 0 {
						vlanIface, err := ensureVlanInterface(br, n.Vlan, n.PreserveDefaultVlan)
						if err != nil {
							return fmt.Errorf("failed to create vlan interface: %v", err)
						}

						if vlanInterface == nil {
							vlanInterface = &current.Interface{
								Name: vlanIface.Attrs().Name,
								Mac:  vlanIface.Attrs().HardwareAddr.String(),
							}
							result.Interfaces = append(result.Interfaces, vlanInterface)
						}

						err = ensureAddr(vlanIface, gws.family, &gw, n.ForceAddress)
						if err != nil {
							return fmt.Errorf("failed to set vlan interface for bridge with addr: %v", err)
						}
					} else {
						err = ensureAddr(br, gws.family, &gw, n.ForceAddress)
						if err != nil {
							return fmt.Errorf("failed to set bridge addr: %v", err)
						}
					}
				}

				if gws.gws != nil {
					if err = enableIPForward(gws.family); err != nil {
						return fmt.Errorf("failed to enable forwarding: %v", err)
					}
				}
			}
		}

		if n.IPMasq {
			chain := utils.FormatChainName(n.Name, args.ContainerID)
			comment := utils.FormatComment(n.Name, args.ContainerID)
			for _, ipc := range result.IPs {
				if err = ip.SetupIPMasq(&ipc.Address, chain, comment); err != nil {
					return err
				}
			}
		}
	} else if !n.DisableContainerInterface {
		if err := netns.Do(func(_ ns.NetNS) error {
			link, err := netlink.LinkByName(args.IfName)
			if err != nil {
				return fmt.Errorf("failed to retrieve link: %v", err)
			}

			// If layer 2 we still need to set the container veth to up
			if err = netlink.LinkSetUp(link); err != nil {
				return fmt.Errorf("failed to set %q up: %v", args.IfName, err)
			}
			return nil
		}); err != nil {
			return err
		}
	}

	hostVeth, err := netlink.LinkByName(hostInterface.Name)
	if err != nil {
		return err
	}

	if !n.DisableContainerInterface {
		// check bridge port state
		retries := []int{0, 50, 500, 1000, 1000}
		for idx, sleep := range retries {
			time.Sleep(time.Duration(sleep) * time.Millisecond)

			hostVeth, err = netlink.LinkByName(hostInterface.Name)
			if err != nil {
				return err
			}
			if hostVeth.Attrs().OperState == netlink.OperUp {
				break
			}

			if idx == len(retries)-1 {
				return fmt.Errorf("bridge port in error state: %s", hostVeth.Attrs().OperState)
			}
		}
	}

	// In certain circumstances, the host-side of the veth may change addrs
	hostInterface.Mac = hostVeth.Attrs().HardwareAddr.String()

	// Refetch the bridge since its MAC address may change when the first
	// veth is added or after its IP address is set
	br, err = bridgeByName(n.BrName)
	if err != nil {
		return err
	}
	brInterface.Mac = br.Attrs().HardwareAddr.String()

	// Return an error requested by testcases, if any
	if debugPostIPAMError != nil {
		return debugPostIPAMError
	}

	// Use incoming DNS settings if provided, otherwise use the
	// settings that were already configured by the IPAM plugin
	if dnsConfSet(n.DNS) {
		result.DNS = n.DNS
	}

	success = true

	return types.PrintResult(result, cniVersion)
}	



func setupVeth(netns ns.NetNS, br *netlink.Bridge, ifName string, mtu int, hairpinMode bool, vlanID int, vlans []int, preserveDefaultVlan bool, mac string) (*current.Interface, *current.Interface, error) {
	contIface := &current.Interface{}
	hostIface := &current.Interface{}

	err := netns.Do(func(hostNS ns.NetNS) error {
		// create the veth pair in the container and move host end into host netns
		hostVeth, containerVeth, err := ip.SetupVeth(ifName, mtu, mac, hostNS)
		if err != nil {
			return err
		}
		contIface.Name = containerVeth.Name
		contIface.Mac = containerVeth.HardwareAddr.String()
		contIface.Sandbox = netns.Path()
		hostIface.Name = hostVeth.Name
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	// need to lookup hostVeth again as its index has changed during ns move
	hostVeth, err := netlink.LinkByName(hostIface.Name)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to lookup %q: %v", hostIface.Name, err)
	}
	hostIface.Mac = hostVeth.Attrs().HardwareAddr.String()

	// connect host veth end to the bridge
	if err := netlink.LinkSetMaster(hostVeth, br); err != nil {
		return nil, nil, fmt.Errorf("failed to connect %q to bridge %v: %v", hostVeth.Attrs().Name, br.Attrs().Name, err)
	}

	// set hairpin mode
	if err = netlink.LinkSetHairpin(hostVeth, hairpinMode); err != nil {
		return nil, nil, fmt.Errorf("failed to setup hairpin mode for %v: %v", hostVeth.Attrs().Name, err)
	}

	if (vlanID != 0 || len(vlans) > 0) && !preserveDefaultVlan {
		err = removeDefaultVlan(hostVeth)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to remove default vlan on interface %q: %v", hostIface.Name, err)
		}
	}

	// Currently bridge CNI only support access port(untagged only) or trunk port(tagged only)
	if vlanID != 0 {
		err = netlink.BridgeVlanAdd(hostVeth, uint16(vlanID), true, true, false, true)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to setup vlan tag on interface %q: %v", hostIface.Name, err)
		}
	}

	for _, v := range vlans {
		err = netlink.BridgeVlanAdd(hostVeth, uint16(v), false, false, false, true)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to setup vlan tag on interface %q: %w", hostIface.Name, err)
		}
	}

	return hostIface, contIface, nil
}


// SetupVeth sets up a pair of virtual ethernet devices.
// Call SetupVeth from inside the container netns.  It will create both veth
// devices and move the host-side veth into the provided hostNS namespace.
// On success, SetupVeth returns (hostVeth, containerVeth, nil)
func SetupVeth(contVethName string, mtu int, contVethMac string, hostNS ns.NetNS) (net.Interface, net.Interface, error) {
	return SetupVethWithName(contVethName, "", mtu, contVethMac, hostNS)
}



// SetupVethWithName sets up a pair of virtual ethernet devices.
// Call SetupVethWithName from inside the container netns.  It will create both veth
// devices and move the host-side veth into the provided hostNS namespace.
// hostVethName: If hostVethName is not specified, the host-side veth name will use a random string.
// On success, SetupVethWithName returns (hostVeth, containerVeth, nil)
func SetupVethWithName(contVethName, hostVethName string, mtu int, contVethMac string, hostNS ns.NetNS) (net.Interface, net.Interface, error) {
	hostVethName, contVeth, err := makeVeth(contVethName, hostVethName, mtu, contVethMac, hostNS)
	if err != nil {
		return net.Interface{}, net.Interface{}, err
	}

	var hostVeth netlink.Link
	err = hostNS.Do(func(_ ns.NetNS) error {
		hostVeth, err = netlink.LinkByName(hostVethName)
		if err != nil {
			return fmt.Errorf("failed to lookup %q in %q: %v", hostVethName, hostNS.Path(), err)
		}

		if err = netlink.LinkSetUp(hostVeth); err != nil {
			return fmt.Errorf("failed to set %q up: %v", hostVethName, err)
		}

		// we want to own the routes for this interface
		_, _ = sysctl.Sysctl(fmt.Sprintf("net/ipv6/conf/%s/accept_ra", hostVethName), "0")
		return nil
	})
	if err != nil {
		return net.Interface{}, net.Interface{}, err
	}
	return ifaceFromNetlinkLink(hostVeth), ifaceFromNetlinkLink(contVeth), nil
}



func makeVeth(name, vethPeerName string, mtu int, mac string, hostNS ns.NetNS) (string, netlink.Link, error) {
	var peerName string
	var veth netlink.Link
	var err error
	for i := 0; i < 10; i++ {
		if vethPeerName != "" {
			peerName = vethPeerName
		} else {
			peerName, err = RandomVethName()
			if err != nil {
				return peerName, nil, err
			}
		}

		veth, err = makeVethPair(name, peerName, mtu, mac, hostNS)
		switch {
		case err == nil:
			return peerName, veth, nil

		case os.IsExist(err):
			if peerExists(peerName) && vethPeerName == "" {
				continue
			}
			return peerName, veth, fmt.Errorf("container veth name (%q) peer provided (%q) already exists", name, peerName)
		default:
			return peerName, veth, fmt.Errorf("failed to make veth pair: %v", err)
		}
	}

	// should really never be hit
	return peerName, nil, fmt.Errorf("failed to find a unique veth name")
}


// makeVethPair is called from within the container's network namespace
func makeVethPair(name, peer string, mtu int, mac string, hostNS ns.NetNS) (netlink.Link, error) {
	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: name,
			MTU:  mtu,
		},
		PeerName:      peer,
		PeerNamespace: netlink.NsFd(int(hostNS.Fd())),
	}
	if mac != "" {
		m, err := net.ParseMAC(mac)
		if err != nil {
			return nil, err
		}
		veth.LinkAttrs.HardwareAddr = m
	}
	if err := netlink.LinkAdd(veth); err != nil {
		return nil, err
	}
	// Re-fetch the container link to get its creation-time parameters, e.g. index and mac
	veth2, err := netlink.LinkByName(name)
	if err != nil {
		netlink.LinkDel(veth) // try and clean up the link if possible.
		return nil, err
	}

	return veth2, nil
}



// LinkAdd adds a new link device. The type and features of the device
// are taken from the parameters in the link object.
// Equivalent to: `ip link add $link`
func LinkAdd(link Link) error {
	return pkgHandle.LinkAdd(link)
}



// LinkAdd adds a new link device. The type and features of the device
// are taken from the parameters in the link object.
// Equivalent to: `ip link add $link`
func (h *Handle) LinkAdd(link Link) error {
	return h.linkModify(link, unix.NLM_F_CREATE|unix.NLM_F_EXCL|unix.NLM_F_ACK)
}

func LinkModify(link Link) error {
	return pkgHandle.LinkModify(link)
}


func (h *Handle) LinkModify(link Link) error {
	return h.linkModify(link, unix.NLM_F_REQUEST|unix.NLM_F_ACK)
}

