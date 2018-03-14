// +build linux

package node

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	networkapi "github.com/openshift/origin/pkg/network/apis/network"
	"github.com/openshift/origin/pkg/network/common"
	"github.com/openshift/origin/pkg/network/node/cniserver"
	"github.com/openshift/origin/pkg/util/netutils"

	"github.com/golang/glog"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	kapi "k8s.io/kubernetes/pkg/apis/core"
	kapiv1 "k8s.io/kubernetes/pkg/apis/core/v1"
	kclientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	kcontainer "k8s.io/kubernetes/pkg/kubelet/container"
	knetwork "k8s.io/kubernetes/pkg/kubelet/network"
	kubehostport "k8s.io/kubernetes/pkg/kubelet/network/hostport"
	kbandwidth "k8s.io/kubernetes/pkg/util/bandwidth"
	utildbus "k8s.io/kubernetes/pkg/util/dbus"
	utiliptables "k8s.io/kubernetes/pkg/util/iptables"
	utilexec "k8s.io/utils/exec"

	"github.com/containernetworking/cni/pkg/invoke"
	cnitypes "github.com/containernetworking/cni/pkg/types"
	cni020 "github.com/containernetworking/cni/pkg/types/020"
	cnicurrent "github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/containernetworking/plugins/pkg/ns"

	"github.com/vishvananda/netlink"
)

const (
	podInterfaceName = knetwork.DefaultInterfaceName
)

type podHandler interface {
	setup(req *cniserver.PodRequest) (cnitypes.Result, *runningPod, error)
	update(req *cniserver.PodRequest) (uint32, error)
	teardown(req *cniserver.PodRequest) error
}

type runningPod struct {
	podPortMapping *kubehostport.PodPortMapping
	vnid           uint32
	ofport         int
}

type podManager struct {
	// Common stuff used for both live and testing code
	podHandler podHandler
	cniServer  *cniserver.CNIServer
	// Request queue for pod operations incoming from the CNIServer
	requests chan (*cniserver.PodRequest)
	// Tracks pod :: IP address for hostport and multicast handling
	runningPods     map[string]*runningPod
	runningPodsLock sync.Mutex

	// Live pod setup/teardown stuff not used in testing code
	kClient kclientset.Interface
	policy  osdnPolicy
	mtu     uint32
	ovs     *ovsController

	enableHostports bool
	// true if hostports have been synced at least once
	hostportsSynced bool
	// true if at least one running pod has a hostport mapping
	activeHostports bool

	// Things only accessed through the processCNIRequests() goroutine
	// and thus can be set from Start()
	ipamConfig     []byte
	hostportSyncer kubehostport.HostportSyncer
}

// Creates a new live podManager; used by node code0
func newPodManager(kClient kclientset.Interface, policy osdnPolicy, mtu uint32, ovs *ovsController, enableHostports bool) *podManager {
	pm := newDefaultPodManager()
	pm.kClient = kClient
	pm.policy = policy
	pm.mtu = mtu
	pm.podHandler = pm
	pm.ovs = ovs
	pm.enableHostports = enableHostports
	return pm
}

// Creates a new basic podManager; used by testcases
func newDefaultPodManager() *podManager {
	return &podManager{
		runningPods: make(map[string]*runningPod),
		requests:    make(chan *cniserver.PodRequest, 20),
	}
}

// Generates a CNI IPAM config from a given node cluster and local subnet that
// CNI 'host-local' IPAM plugin will use to create an IP address lease for the
// container
func getIPAMConfig(clusterNetworks []common.ClusterNetwork, localSubnet string) ([]byte, error) {
	nodeNet, err := cnitypes.ParseCIDR(localSubnet)
	if err != nil {
		return nil, fmt.Errorf("error parsing node network '%s': %v", localSubnet, err)
	}

	type hostLocalIPAM struct {
		Type   string           `json:"type"`
		Subnet cnitypes.IPNet   `json:"subnet"`
		Routes []cnitypes.Route `json:"routes"`
	}

	type cniNetworkConfig struct {
		CNIVersion string         `json:"cniVersion"`
		Name       string         `json:"name"`
		Type       string         `json:"type"`
		IPAM       *hostLocalIPAM `json:"ipam"`
	}

	_, mcnet, _ := net.ParseCIDR("224.0.0.0/4")

	routes := []cnitypes.Route{
		{
			//Default route
			Dst: net.IPNet{
				IP:   net.IPv4zero,
				Mask: net.IPMask(net.IPv4zero),
			},
			GW: netutils.GenerateDefaultGateway(nodeNet),
		},
		{
			//Multicast
			Dst: *mcnet,
		},
	}

	for _, cn := range clusterNetworks {
		routes = append(routes, cnitypes.Route{Dst: *cn.ClusterCIDR})
	}

	return json.Marshal(&cniNetworkConfig{
		// TODO: update to 0.3.0 spec
		CNIVersion: "0.2.0",
		Name:       "openshift-sdn",
		Type:       "openshift-sdn",
		IPAM: &hostLocalIPAM{
			Type: "host-local",
			Subnet: cnitypes.IPNet{
				IP:   nodeNet.IP,
				Mask: nodeNet.Mask,
			},
			Routes: routes,
		},
	})
}

// Start the CNI server and start processing requests from it
func (m *podManager) Start(socketPath string, localSubnetCIDR string, clusterNetworks []common.ClusterNetwork) error {
	if m.enableHostports {
		iptInterface := utiliptables.New(utilexec.New(), utildbus.New(), utiliptables.ProtocolIpv4)
		m.hostportSyncer = kubehostport.NewHostportSyncer(iptInterface)
	}

	var err error
	if m.ipamConfig, err = getIPAMConfig(clusterNetworks, localSubnetCIDR); err != nil {
		return err
	}

	go m.processCNIRequests()

	m.cniServer = cniserver.NewCNIServer(socketPath)
	return m.cniServer.Start(m.handleCNIRequest)
}

// Returns a key for use with the runningPods map
func getPodKey(request *cniserver.PodRequest) string {
	return fmt.Sprintf("%s/%s", request.PodNamespace, request.PodName)
}

func (m *podManager) getPod(request *cniserver.PodRequest) *kubehostport.PodPortMapping {
	if pod := m.runningPods[getPodKey(request)]; pod != nil {
		return pod.podPortMapping
	}
	return nil
}

func hasHostPorts(pod *kubehostport.PodPortMapping) bool {
	for _, mapping := range pod.PortMappings {
		if mapping.HostPort != 0 {
			return true
		}
	}
	return false
}

// Return a list of Kubernetes RunningPod objects for hostport operations
func (m *podManager) shouldSyncHostports(newPod *kubehostport.PodPortMapping) []*kubehostport.PodPortMapping {
	if m.hostportSyncer == nil {
		return nil
	}

	newActiveHostports := false
	mappings := make([]*kubehostport.PodPortMapping, 0)
	for _, runningPod := range m.runningPods {
		mappings = append(mappings, runningPod.podPortMapping)
		if !newActiveHostports && hasHostPorts(runningPod.podPortMapping) {
			newActiveHostports = true
		}
	}
	if newPod != nil && hasHostPorts(newPod) {
		newActiveHostports = true
	}

	// Sync the first time a pod is started (to clear out stale mappings
	// if kubelet crashed), or when there are any/will be active hostports.
	// Otherwise don't bother.
	if !m.hostportsSynced || m.activeHostports || newActiveHostports {
		m.hostportsSynced = true
		m.activeHostports = newActiveHostports
		return mappings
	}

	return nil
}

// Add a request to the podManager CNI request queue
func (m *podManager) addRequest(request *cniserver.PodRequest) {
	m.requests <- request
}

// Wait for and return the result of a pod request
func (m *podManager) waitRequest(request *cniserver.PodRequest) *cniserver.PodResult {
	return <-request.Result
}

// Enqueue incoming pod requests from the CNI server, wait on the result,
// and return that result to the CNI client
func (m *podManager) handleCNIRequest(request *cniserver.PodRequest) ([]byte, error) {
	glog.V(5).Infof("Dispatching pod network request %v", request)
	m.addRequest(request)
	result := m.waitRequest(request)
	glog.V(5).Infof("Returning pod network request %v, result %s err %v", request, string(result.Response), result.Err)
	return result.Response, result.Err
}

func (m *podManager) updateLocalMulticastRulesWithLock(vnid uint32) {
	var ofports []int
	enabled := m.policy.GetMulticastEnabled(vnid)
	if enabled {
		for _, pod := range m.runningPods {
			if pod.vnid == vnid {
				ofports = append(ofports, pod.ofport)
			}
		}
	}

	if err := m.ovs.UpdateLocalMulticastFlows(vnid, enabled, ofports); err != nil {
		glog.Errorf("Error updating OVS multicast flows for VNID %d: %v", vnid, err)

	}
}

// Update multicast OVS rules for the given vnid
func (m *podManager) UpdateLocalMulticastRules(vnid uint32) {
	m.runningPodsLock.Lock()
	defer m.runningPodsLock.Unlock()
	m.updateLocalMulticastRulesWithLock(vnid)
}

// Process all CNI requests from the request queue serially.  Our OVS interaction
// and scripts currently cannot run in parallel, and doing so greatly complicates
// setup/teardown logic
func (m *podManager) processCNIRequests() {
	for request := range m.requests {
		glog.V(5).Infof("Processing pod network request %v", request)
		result := m.processRequest(request)
		glog.V(5).Infof("Processed pod network request %v, result %s err %v", request, string(result.Response), result.Err)
		request.Result <- result
	}
	panic("stopped processing CNI pod requests!")
}

func (m *podManager) processRequest(request *cniserver.PodRequest) *cniserver.PodResult {
	m.runningPodsLock.Lock()
	defer m.runningPodsLock.Unlock()

	pk := getPodKey(request)
	result := &cniserver.PodResult{}
	switch request.Command {
	case cniserver.CNI_ADD:
		ipamResult, runningPod, err := m.podHandler.setup(request)
		if ipamResult != nil {
			result.Response, err = json.Marshal(ipamResult)
			if err == nil {
				m.runningPods[pk] = runningPod
				if m.ovs != nil {
					m.updateLocalMulticastRulesWithLock(runningPod.vnid)
				}
			}
		}
		if err != nil {
			PodOperationsErrors.WithLabelValues(PodOperationSetup).Inc()
			result.Err = err
		}
	case cniserver.CNI_UPDATE:
		vnid, err := m.podHandler.update(request)
		if err == nil {
			if runningPod, exists := m.runningPods[pk]; exists {
				runningPod.vnid = vnid
			}
		}
		result.Err = err
	case cniserver.CNI_DEL:
		if runningPod, exists := m.runningPods[pk]; exists {
			delete(m.runningPods, pk)
			if m.ovs != nil {
				m.updateLocalMulticastRulesWithLock(runningPod.vnid)
			}
		}
		result.Err = m.podHandler.teardown(request)
		if result.Err != nil {
			PodOperationsErrors.WithLabelValues(PodOperationTeardown).Inc()
		}
	default:
		result.Err = fmt.Errorf("unhandled CNI request %v", request.Command)
	}
	return result
}

// For a given container, returns host veth name, container veth MAC, and pod IP
func getVethInfo(netns, containerIfname string) (string, string, string, error) {
	var (
		peerIfindex int
		contVeth    netlink.Link
		err         error
		podIP       string
	)

	containerNs, err := ns.GetNS(netns)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get container netns: %v", err)
	}
	defer containerNs.Close()

	err = containerNs.Do(func(ns.NetNS) error {
		contVeth, err = netlink.LinkByName(containerIfname)
		if err != nil {
			return err
		}
		peerIfindex = contVeth.Attrs().ParentIndex

		addrs, err := netlink.AddrList(contVeth, netlink.FAMILY_V4)
		if err != nil {
			return fmt.Errorf("failed to get container IP addresses: %v", err)
		}
		if len(addrs) == 0 {
			return fmt.Errorf("container had no addresses")
		}
		podIP = addrs[0].IP.String()

		return nil
	})
	if err != nil {
		return "", "", "", fmt.Errorf("failed to inspect container interface: %v", err)
	}

	hostVeth, err := netlink.LinkByIndex(peerIfindex)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get host veth: %v", err)
	}

	return hostVeth.Attrs().Name, contVeth.Attrs().HardwareAddr.String(), podIP, nil
}

// Adds a macvlan interface to a container, if requested, for use with the egress router feature
func maybeAddMacvlan(pod *kapi.Pod, netns string) error {
	annotation, ok := pod.Annotations[networkapi.AssignMacvlanAnnotation]
	if !ok || annotation == "false" {
		return nil
	}

	privileged := false
	for _, container := range append(pod.Spec.Containers, pod.Spec.InitContainers...) {
		if container.SecurityContext != nil && container.SecurityContext.Privileged != nil && *container.SecurityContext.Privileged {
			privileged = true
			break
		}
	}
	if !privileged {
		return fmt.Errorf("pod has %q annotation but is not privileged", networkapi.AssignMacvlanAnnotation)
	}

	var iface netlink.Link
	var err error
	if annotation == "true" {
		// Find interface with the default route
		routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
		if err != nil {
			return fmt.Errorf("failed to read routes: %v", err)
		}

		for _, r := range routes {
			if r.Dst == nil {
				iface, err = netlink.LinkByIndex(r.LinkIndex)
				if err != nil {
					return fmt.Errorf("failed to get default route interface: %v", err)
				}
			}
		}
		if iface == nil {
			return fmt.Errorf("failed to find default route interface")
		}
	} else {
		iface, err = netlink.LinkByName(annotation)
		if err != nil {
			return fmt.Errorf("pod annotation %q is neither 'true' nor the name of a local network interface", networkapi.AssignMacvlanAnnotation)
		}
	}

	podNs, err := ns.GetNS(netns)
	if err != nil {
		return fmt.Errorf("could not open netns %q", netns)
	}
	defer podNs.Close()

	err = netlink.LinkAdd(&netlink.Macvlan{
		LinkAttrs: netlink.LinkAttrs{
			MTU:         iface.Attrs().MTU,
			Name:        "macvlan0",
			ParentIndex: iface.Attrs().Index,
			Namespace:   netlink.NsFd(podNs.Fd()),
		},
		Mode: netlink.MACVLAN_MODE_PRIVATE,
	})
	if err != nil {
		return fmt.Errorf("failed to create macvlan interface: %v", err)
	}
	return podNs.Do(func(netns ns.NetNS) error {
		l, err := netlink.LinkByName("macvlan0")
		if err != nil {
			return fmt.Errorf("failed to find macvlan interface: %v", err)
		}
		err = netlink.LinkSetUp(l)
		if err != nil {
			return fmt.Errorf("failed to set macvlan interface up: %v", err)
		}
		return nil
	})
}

func createIPAMArgs(netnsPath string, action cniserver.CNICommand, id string) *invoke.Args {
	return &invoke.Args{
		Command:     string(action),
		ContainerID: id,
		NetNS:       netnsPath,
		IfName:      podInterfaceName,
		Path:        "/opt/cni/bin",
	}
}

// Run CNI IPAM allocation for the container and return the allocated IP address
func (m *podManager) ipamAdd(netnsPath string, id string) (*cni020.Result, net.IP, error) {
	if netnsPath == "" {
		return nil, nil, fmt.Errorf("netns required for CNI_ADD")
	}

	args := createIPAMArgs(netnsPath, cniserver.CNI_ADD, id)
	r, err := invoke.ExecPluginWithResult("/opt/cni/bin/host-local", m.ipamConfig, args)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to run CNI IPAM ADD: %v", err)
	}

	// We gave the IPAM plugin 0.2.0 config, so the plugin must return a 0.2.0 result
	result, err := cni020.GetResult(r)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CNI IPAM ADD result: %v", err)
	}
	if result.IP4 == nil {
		return nil, nil, fmt.Errorf("failed to obtain IP address from CNI IPAM")
	}

	return result, result.IP4.IP.IP, nil
}

// Run CNI IPAM release for the container
func (m *podManager) ipamDel(id string) error {
	args := createIPAMArgs("", cniserver.CNI_DEL, id)
	err := invoke.ExecPluginWithoutResult("/opt/cni/bin/host-local", m.ipamConfig, args)
	if err != nil {
		return fmt.Errorf("failed to run CNI IPAM DEL: %v", err)
	}
	return nil
}

func setupPodBandwidth(ovs *ovsController, pod *kapi.Pod, hostVeth string) error {
	ingressVal, egressVal, err := kbandwidth.ExtractPodBandwidthResources(pod.Annotations)
	if err != nil {
		return fmt.Errorf("failed to parse pod bandwidth: %v", err)
	}

	ingressBPS := int64(-1)
	egressBPS := int64(-1)
	if ingressVal != nil {
		ingressBPS = ingressVal.Value()

		l, err := netlink.LinkByName(hostVeth)
		if err != nil {
			return fmt.Errorf("failed to find host veth interface %s: %v", hostVeth, err)
		}
		err = netlink.LinkSetTxQLen(l, 1000)
		if err != nil {
			return fmt.Errorf("failed to set host veth txqlen: %v", err)
		}
	}
	if egressVal != nil {
		egressBPS = egressVal.Value()
	}

	return ovs.SetPodBandwidth(hostVeth, ingressBPS, egressBPS)
}

func vnidToString(vnid uint32) string {
	return strconv.FormatUint(uint64(vnid), 10)
}

// podIsExited returns true if the pod is exited (all containers inside are exited).
func podIsExited(p *kcontainer.Pod) bool {
	for _, c := range p.Containers {
		if c.State != kcontainer.ContainerStateExited {
			return false
		}
	}
	return true
}

// Set up all networking (host/container veth, OVS flows, IPAM, loopback, etc)
func (m *podManager) setup(req *cniserver.PodRequest) (cnitypes.Result, *runningPod, error) {
	defer PodOperationsLatency.WithLabelValues(PodOperationSetup).Observe(sinceInMicroseconds(time.Now()))

	pod, err := m.kClient.Core().Pods(req.PodNamespace).Get(req.PodName, metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}

	ipamResult, podIP, err := m.ipamAdd(req.Netns, req.SandboxID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to run IPAM for %v: %v", req.SandboxID, err)
	}

	// Release any IPAM allocations and hostports if the setup failed
	var success bool
	defer func() {
		if !success {
			m.ipamDel(req.SandboxID)
			if mappings := m.shouldSyncHostports(nil); mappings != nil {
				if err := m.hostportSyncer.SyncHostports(Tun0, mappings); err != nil {
					glog.Warningf("failed syncing hostports: %v", err)
				}
			}
		}
	}()

	// Open any hostports the pod wants
	var v1Pod v1.Pod
	if err := kapiv1.Convert_core_Pod_To_v1_Pod(pod, &v1Pod, nil); err != nil {
		return nil, nil, err
	}
	podPortMapping := kubehostport.ConstructPodPortMapping(&v1Pod, podIP)
	if mappings := m.shouldSyncHostports(podPortMapping); mappings != nil {
		if err := m.hostportSyncer.OpenPodHostportsAndSync(podPortMapping, Tun0, mappings); err != nil {
			return nil, nil, err
		}
	}

	var hostVethName, contVethMac string
	err = ns.WithNetNSPath(req.Netns, func(hostNS ns.NetNS) error {
		hostVeth, contVeth, err := ip.SetupVeth(podInterfaceName, int(m.mtu), hostNS)
		if err != nil {
			return fmt.Errorf("failed to create container veth: %v", err)
		}
		// Force a consistent MAC address based on the IP address
		if err := ip.SetHWAddrByIP(podInterfaceName, podIP, nil); err != nil {
			return fmt.Errorf("failed to set pod interface MAC address: %v", err)
		}
		// refetch to get hardware address and other properties
		tmp, err := net.InterfaceByIndex(contVeth.Index)
		if err != nil {
			return fmt.Errorf("failed to fetch container veth: %v", err)
		}
		contVeth = *tmp

		// Clear out gateway to prevent ConfigureIface from adding the cluster
		// subnet via the gateway
		ipamResult.IP4.Gateway = nil
		result030, err := cnicurrent.NewResultFromResult(ipamResult)
		if err != nil {
			return fmt.Errorf("failed to convert IPAM: %v", err)
		}
		// Add a sandbox interface record which ConfigureInterface expects.
		// The only interface we report is the pod interface.
		result030.Interfaces = []*cnicurrent.Interface{
			{
				Name:    podInterfaceName,
				Mac:     contVeth.HardwareAddr.String(),
				Sandbox: req.Netns,
			},
		}
		intPtr := 0
		result030.IPs[0].Interface = &intPtr

		if err = ipam.ConfigureIface(podInterfaceName, result030); err != nil {
			return fmt.Errorf("failed to configure container IPAM: %v", err)
		}

		lo, err := netlink.LinkByName("lo")
		if err == nil {
			err = netlink.LinkSetUp(lo)
		}
		if err != nil {
			return fmt.Errorf("failed to configure container loopback: %v", err)
		}

		hostVethName = hostVeth.Name
		contVethMac = contVeth.HardwareAddr.String()
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	vnid, err := m.policy.GetVNID(req.PodNamespace)
	if err != nil {
		return nil, nil, err
	}

	if err := maybeAddMacvlan(pod, req.Netns); err != nil {
		return nil, nil, err
	}

	ofport, err := m.ovs.SetUpPod(hostVethName, podIP.String(), contVethMac, req.SandboxID, vnid)
	if err != nil {
		return nil, nil, err
	}
	if err := setupPodBandwidth(m.ovs, pod, hostVethName); err != nil {
		return nil, nil, err
	}

	m.policy.EnsureVNIDRules(vnid)
	success = true
	return ipamResult, &runningPod{podPortMapping: podPortMapping, vnid: vnid, ofport: ofport}, nil
}

// Update OVS flows when something (like the pod's namespace VNID) changes
func (m *podManager) update(req *cniserver.PodRequest) (uint32, error) {
	vnid, err := m.policy.GetVNID(req.PodNamespace)
	if err != nil {
		return 0, err
	}

	if err := m.ovs.UpdatePod(req.SandboxID, vnid); err != nil {
		return 0, err
	}
	return vnid, nil
}

// Clean up all pod networking (clear OVS flows, release IPAM lease, remove host/container veth)
func (m *podManager) teardown(req *cniserver.PodRequest) error {
	defer PodOperationsLatency.WithLabelValues(PodOperationTeardown).Observe(sinceInMicroseconds(time.Now()))

	netnsValid := true
	if err := ns.IsNSorErr(req.Netns); err != nil {
		if _, ok := err.(ns.NSPathNotExistErr); ok {
			glog.V(3).Infof("teardown called on already-destroyed pod %s/%s; only cleaning up IPAM", req.PodNamespace, req.PodName)
			netnsValid = false
		}
	}

	errList := []error{}
	if netnsValid {
		hostVethName, _, podIP, err := getVethInfo(req.Netns, podInterfaceName)
		if err != nil {
			return err
		}

		if err := m.ovs.TearDownPod(hostVethName, podIP, req.SandboxID); err != nil {
			errList = append(errList, err)
		}
	}

	if err := m.ipamDel(req.SandboxID); err != nil {
		errList = append(errList, err)
	}

	if mappings := m.shouldSyncHostports(nil); mappings != nil {
		if err := m.hostportSyncer.SyncHostports(Tun0, mappings); err != nil {
			errList = append(errList, err)
		}
	}

	return kerrors.NewAggregate(errList)
}
