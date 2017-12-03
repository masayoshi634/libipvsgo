package libipvs

import (
	"fmt"
	"syscall"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netlink/nl"
)

var (
	native = nl.NativeEndian()
)

type IPVSHandler struct {
	familyID int
}

func NewIPVSHandler() (*IPVSHandler, error) {
	ipvs := &IPVSHandler{}
	ipvs.getIPVSFamilyID()
	return ipvs, nil
}

func (h *IPVSHandler) getIPVSFamilyID() error {
	var err error
	ipvsFamily, err := netlink.GenlFamilyGet("IPVS")
	if err != nil {
		return err
	}
	h.familyID = int(ipvsFamily.ID)
	return nil
}

func (h *IPVSHandler) sendRequest(cmd uint8, si *ServiceEntry, di *DestinationEntry) ([][]byte, error) {
	var err error
	var serviceAttr nl.NetlinkRequestData
	var destinationAttr nl.NetlinkRequestData

	req := nl.NewNetlinkRequest(int(h.familyID), syscall.NLM_F_ACK)
	req.AddData(&nl.Genlmsg{Command: cmd, Version: 1})

	if si != nil {
		serviceAttr, err = si.Serialize()
		if err != nil {
			return nil, err
		}
	}

	if di != nil {
		destinationAttr, err = di.Serialize()
		if err != nil {
			return nil, err
		}
	}

	switch cmd {
	case IPVS_CMD_FLUSH:
	case IPVS_CMD_GET_SERVICE:
		req.Flags |= syscall.NLM_F_DUMP
	case IPVS_CMD_GET_DEST:
		req.Flags |= syscall.NLM_F_DUMP
		req.AddData(serviceAttr)
	case IPVS_CMD_NEW_SERVICE:
		req.AddData(serviceAttr)
	case IPVS_CMD_SET_SERVICE:
		req.AddData(serviceAttr)
	case IPVS_CMD_DEL_SERVICE:
		req.AddData(serviceAttr)
	case IPVS_CMD_NEW_DEST:
		req.AddData(serviceAttr)
		req.AddData(destinationAttr)
	case IPVS_CMD_SET_DEST:
		req.AddData(serviceAttr)
		req.AddData(destinationAttr)
	case IPVS_CMD_DEL_DEST:
		req.AddData(serviceAttr)
		req.AddData(destinationAttr)
	}

	resp, err := req.Execute(syscall.NETLINK_GENERIC, 0)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (h *IPVSHandler) parseIPVSServiceMessage(attrs [][]syscall.NetlinkRouteAttr) ([]*ServiceEntry, error) {
	var services []*ServiceEntry
	for _, ipvsAttrs := range attrs {
		s, err := assembleServiceInterface(ipvsAttrs)
		if err != nil {
			return nil, err
		}
		services = append(services, s)
	}
	return services, nil
}

func (h *IPVSHandler) parseIPVSDestinationMessage(attrs [][]syscall.NetlinkRouteAttr) ([]*DestinationEntry, error) {
	var destinations []*DestinationEntry
	for _, ipvsAttrs := range attrs {
		d, err := assembleDestinationInterface(ipvsAttrs)
		if err != nil {
			return nil, err
		}
		destinations = append(destinations, d)
	}
	return destinations, nil
}

func (h *IPVSHandler) parseGenlHeaders(msgs [][]byte) ([][]syscall.NetlinkRouteAttr, error) {
	var ipvsAttrsList [][]syscall.NetlinkRouteAttr
	for _, msg := range msgs {
		NetLinkAttrs, err := nl.ParseRouteAttr(msg[nl.SizeofGenlmsg:])
		if err != nil {
			return nil, err
		}
		if len(NetLinkAttrs) == 0 {
			return nil, fmt.Errorf("invalid netlink message")
		}

		ipvsAttrs, err := nl.ParseRouteAttr(NetLinkAttrs[0].Value)
		if err != nil {
			return nil, err
		}
		ipvsAttrsList = append(ipvsAttrsList, ipvsAttrs)
	}
	return ipvsAttrsList, nil
}

func (h *IPVSHandler) GetAllEntry() ([]*Entry, error) {
	var err error
	var entries []*Entry
	services, err := h.GetServices()
	if err != nil {
		return nil, err
	}

	for _, s := range services {
		d, err := h.GetDestinations(s)
		if err != nil {
			return nil, err
		}
		entries = append(entries, &Entry{Service: s, Destinations: d})
	}

	return entries, nil
}

func (h *IPVSHandler) IsRegisteredService(vip string, port int, protocol string) (bool, error) {
	cmd := IPVS_CMD_GET_SERVICE
	si := &ServiceEntry{Address: vip, Protocol: protocol, Port: port}

	msgs, err := h.sendRequest(cmd, si, nil)
	if err != nil {
		return false, err
	}

	ipvsAttrs, err := h.parseGenlHeaders(msgs)
	if err != nil {
		return false, err
	}
	services, err := h.parseIPVSServiceMessage(ipvsAttrs)
	if err != nil {
		return false, err
	}
	if len(services) < 1 {
		return false, nil
	}
	return true, nil
}

func (h *IPVSHandler) GetService(vip string, port int, protocol string) (*ServiceEntry, error) {
	cmd := IPVS_CMD_GET_SERVICE
	si := &ServiceEntry{Address: vip, Protocol: protocol, Port: port}
	msgs, err := h.sendRequest(cmd, si, nil)
	if err != nil {
		return nil, err
	}

	ipvsAttrs, err := h.parseGenlHeaders(msgs)
	if err != nil {
		return nil, err
	}
	services, err := h.parseIPVSServiceMessage(ipvsAttrs)
	if err != nil {
		return nil, err
	}
	if len(services) < 1 {
		return nil, fmt.Errorf("not found entry %v", services)
	} else if len(services) > 1 {
		return nil, fmt.Errorf("found duplicate entry %v", services)
	}
	return services[0], nil
}

func (h *IPVSHandler) GetServices() ([]*ServiceEntry, error) {
	var err error
	cmd := IPVS_CMD_GET_SERVICE
	msgs, err := h.sendRequest(cmd, nil, nil)
	if err != nil {
		return nil, err
	}

	ipvsAttrs, err := h.parseGenlHeaders(msgs)
	if err != nil {
		return nil, err
	}
	services, err := h.parseIPVSServiceMessage(ipvsAttrs)
	return services, nil
}

func (h *IPVSHandler) UpdateService(ip string, port int, protocol string, schedName string) error {
	var err error
	cmd := IPVS_CMD_SET_SERVICE
	si := &ServiceEntry{Address: ip, Protocol: protocol, Port: port, SchedName: schedName}
	_, err = h.sendRequest(cmd, si, nil)
	if err != nil {
		return err
	}
	return nil
}

func (h *IPVSHandler) DeleteService(vip string, port int, protocol string) error {
	cmd := IPVS_CMD_DEL_SERVICE
	si, err := h.GetService(vip, port, protocol)
	if err != nil {
		return err
	}

	_, err = h.sendRequest(cmd, si, nil)
	if err != nil {
		return err
	}
	return nil
}

func (h *IPVSHandler) Flush() error {
	cmd := IPVS_CMD_FLUSH
	_, err := h.sendRequest(cmd, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

func (h *IPVSHandler) AddService(vip string, port int, protocol string, schedName string) error {
	var err error
	cmd := IPVS_CMD_NEW_SERVICE

	si := &ServiceEntry{Address: vip, Protocol: protocol, Port: port, SchedName: schedName}

	_, err = h.sendRequest(cmd, si, nil)
	if err != nil {
		return err
	}
	return nil
}

func (h *IPVSHandler) AddDestination(si *ServiceEntry, ip string, port int, weight int, method string) error {
	var err error
	cmd := IPVS_CMD_NEW_DEST
	di := &DestinationEntry{Address: ip, Port: port, Weight: weight, Method: method}

	_, err = h.sendRequest(cmd, si, di)
	if err != nil {
		return err
	}
	return nil
}

func (h *IPVSHandler) UpdateDestination(si *ServiceEntry, ip string, port int, weight int, method string) error {
	var err error
	cmd := IPVS_CMD_SET_DEST
	di := &DestinationEntry{Address: ip, Port: port, Weight: weight, Method: method}

	_, err = h.sendRequest(cmd, si, di)
	if err != nil {
		return err
	}
	return nil
}
func (h *IPVSHandler) GetDestination(si *ServiceEntry, rip string, rport int) (*DestinationEntry, error) {
	var err error
	cmd := IPVS_CMD_GET_DEST
	di := &DestinationEntry{Address: rip, Port: rport}

	msgs, err := h.sendRequest(cmd, si, di)
	if err != nil {
		return nil, err
	}

	ipvsAttrs, err := h.parseGenlHeaders(msgs)
	if err != nil {
		return nil, err
	}
	destinations, err := h.parseIPVSDestinationMessage(ipvsAttrs)
	if err != nil {
		return nil, err
	}
	if len(destinations) < 1 {
		return nil, fmt.Errorf("Not found destination")
	} else if len(destinations) > 1 {
		return nil, fmt.Errorf("found duplication destination")
	}
	return destinations[0], nil
}

func (h *IPVSHandler) GetDestinations(si *ServiceEntry) ([]*DestinationEntry, error) {
	var err error
	cmd := IPVS_CMD_GET_DEST

	msgs, err := h.sendRequest(cmd, si, nil)
	if err != nil {
		return nil, err
	}

	ipvsAttrs, err := h.parseGenlHeaders(msgs)
	if err != nil {
		return nil, err
	}
	destinations, err := h.parseIPVSDestinationMessage(ipvsAttrs)
	if err != nil {
		return nil, err
	}
	return destinations, nil
}

func (h *IPVSHandler) DeleteDestination(vip string, vport int, rip string, rport int, protocol string) error {
	var err error
	cmd := IPVS_CMD_DEL_DEST
	si, err := h.GetService(vip, vport, protocol)
	if err != nil {
		return err
	}

	di, err := h.GetDestination(si, rip, rport)
	if err != nil {
		return err
	}
	_, err = h.sendRequest(cmd, si, di)
	if err != nil {
		return err
	}
	return nil
}
