package libipvs

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"syscall"
	"unsafe"

	"github.com/vishvananda/netlink/nl"
)

var (
	Name     string
	Version  string
	Revision string
)

type Entry struct {
	Service      *ServiceEntry
	Destinations []*DestinationEntry
}

type ServiceEntry struct {
	Address  string `json:"vip"`
	Protocol string `json:"protocol"`
	Port     int    `json:"port"`
	FWMark   int    `json:"fwmark"`

	SchedName     string `json:"scheduler"`
	Flags         int    `json:"flags"`
	Timeout       int    `json:"timeout"`
	Netmask       int    `json:"netmask"`
	AddressFamily string `json:"addressfamily"`
	PEName        string `json:"pename"`
	Stats         Stats
}

func (s *ServiceEntry) Serialize() (nl.NetlinkRequestData, error) {
	ip := net.ParseIP(s.Address)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address %s", s.Address)
	}

	var protocol uint16
	switch s.Protocol {
	case "TCP", "tcp", "Tcp":
		protocol = syscall.IPPROTO_TCP //0x6
	case "UDP", "udp", "Udp":
		protocol = syscall.IPPROTO_UDP //0x11
	case "SCTP", "sctp", "Sctp":
		protocol = syscall.IPPROTO_SCTP //0x84
	default:
		return nil, fmt.Errorf("not support protocol %s", s.Protocol)
	}

	var addressFamily uint16
	addressFamily = syscall.AF_INET // 0x2
	if ip.To4() == nil {
		addressFamily = syscall.AF_INET6 //0xa
	}
	switch s.SchedName {
	case "", "rr", "wrr", "lc", "wlc", "lblc", "able", "lblcr", "dh", "sh", "sed", "nq":
	default:
		return nil, fmt.Errorf("not support scheduler %s", s.SchedName)
	}

	cmdAttrService := nl.NewRtAttr(IPVS_CMD_ATTR_SERVICE, nil)
	nl.NewRtAttrChild(cmdAttrService, IPVS_SVC_ATTR_AF, nl.Uint16Attr(addressFamily))

	if s.FWMark == 0 {
		portBuf := new(bytes.Buffer)
		binary.Write(portBuf, binary.BigEndian, uint16(s.Port))
		nl.NewRtAttrChild(cmdAttrService, IPVS_SVC_ATTR_PORT, portBuf.Bytes())
		nl.NewRtAttrChild(cmdAttrService, IPVS_SVC_ATTR_PROTOCOL, nl.Uint16Attr(protocol))
		switch addressFamily {
		case syscall.AF_INET:
			nl.NewRtAttrChild(cmdAttrService, IPVS_SVC_ATTR_ADDR, ip.To4())
		case syscall.AF_INET6:
			nl.NewRtAttrChild(cmdAttrService, IPVS_SVC_ATTR_ADDR, ip)
		}
	} else {
		nl.NewRtAttrChild(cmdAttrService, IPVS_SVC_ATTR_FWMARK, nl.Uint32Attr(uint32(s.FWMark)))
	}

	nl.NewRtAttrChild(cmdAttrService, IPVS_SVC_ATTR_SCHED_NAME, nl.ZeroTerminated(s.SchedName))
	if s.PEName != "" {
		nl.NewRtAttrChild(cmdAttrService, IPVS_SVC_ATTR_PE_NAME, nl.ZeroTerminated(s.PEName))
	}
	f := &ipvsFlags{
		flags: uint32(s.Flags),
		mask:  0xFFFFFFFF,
	}
	nl.NewRtAttrChild(cmdAttrService, IPVS_SVC_ATTR_FLAGS, f.Serialize())
	nl.NewRtAttrChild(cmdAttrService, IPVS_SVC_ATTR_TIMEOUT, nl.Uint32Attr(uint32(s.Timeout)))
	nl.NewRtAttrChild(cmdAttrService, IPVS_SVC_ATTR_NETMASK, nl.Uint32Attr(uint32(s.Netmask)))
	return cmdAttrService, nil
}

type DestinationEntry struct {
	Address        string `json:"RIP"`
	Port           int    `json:"port"`
	Weight         int    `json:"weight"`
	Method         string `json:"method"`
	AddressFamily  string `json:"addressfamily"`
	UpperThreshold int
	LowerThreshold int

	ActiveConnections   uint32
	InActiveConnections uint32
	PersistConnections  uint32
	Stats               Stats
}

func (d *DestinationEntry) Serialize() (nl.NetlinkRequestData, error) {
	ip := net.ParseIP(d.Address)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address %s", d.Address)
	}

	var method uint32
	switch d.Method {
	case "NAT":
		method = IP_VS_CONN_F_MASQ
	case "DR":
		method = IP_VS_CONN_F_DROUTE
	case "TUN":
		method = IP_VS_CONN_F_TUNNEL
	default:
		return nil, errors.New("not support method " + d.Method)
	}

	var addressFamily uint16
	addressFamily = syscall.AF_INET
	if ip.To4() == nil {
		addressFamily = syscall.AF_INET6
	}

	cmdAttrDest := nl.NewRtAttr(IPVS_CMD_ATTR_DEST, nil)

	switch addressFamily {
	case syscall.AF_INET:
		nl.NewRtAttrChild(cmdAttrDest, IPVS_DEST_ATTR_ADDR, ip.To4())
	case syscall.AF_INET6:
		nl.NewRtAttrChild(cmdAttrDest, IPVS_DEST_ATTR_ADDR, ip)
	}

	portBuf := new(bytes.Buffer)
	binary.Write(portBuf, binary.BigEndian, uint16(d.Port))
	nl.NewRtAttrChild(cmdAttrDest, IPVS_DEST_ATTR_PORT, portBuf.Bytes())

	nl.NewRtAttrChild(cmdAttrDest, IPVS_DEST_ATTR_FWD_METHOD, nl.Uint32Attr(method&IP_VS_CONN_F_FWD_MASK))
	nl.NewRtAttrChild(cmdAttrDest, IPVS_DEST_ATTR_WEIGHT, nl.Uint32Attr(uint32(d.Weight)))
	nl.NewRtAttrChild(cmdAttrDest, IPVS_DEST_ATTR_U_THRESH, nl.Uint32Attr(uint32(d.UpperThreshold)))
	nl.NewRtAttrChild(cmdAttrDest, IPVS_DEST_ATTR_L_THRESH, nl.Uint32Attr(uint32(d.LowerThreshold)))

	return cmdAttrDest, nil
}

// Stats defines an IPVS service statistics
type Stats struct {
	Connections uint32 //IPVS_STATS_ATTR_CONNS
	PacketsIn   uint32 //IPVS_STATS_ATTR_INPKTS
	PacketsOut  uint32 //IPVS_STATS_ATTR_OUTPKTS
	BytesIn     uint64 //IPVS_STATS_ATTR_INBYTES
	BytesOut    uint64 //IPVS_STATS_ATTR_OUTBYTES
	CPS         uint32 //IPVS_STATS_ATTR_CPS
	PPSIn       uint32 //IPVS_STATS_ATTR_INPPS
	PPSOut      uint32 //IPVS_STATS_ATTR_OUTPPS
	BPSIn       uint32 //IPVS_STATS_ATTR_INBPS
	BPSOut      uint32 //IPVS_STATS_ATTR_OUTBPS
}

func assembleServiceInterface(attrs []syscall.NetlinkRouteAttr) (*ServiceEntry, error) {
	var s ServiceEntry
	var addr []byte

	for _, attr := range attrs {

		attrType := int(attr.Attr.Type)

		switch attrType {
		case IPVS_SVC_ATTR_AF:
			switch native.Uint16(attr.Value) {
			case syscall.AF_INET:
				s.AddressFamily = "IPv4"
			case syscall.AF_INET6:
				s.AddressFamily = "IPv6"
			}
		case IPVS_SVC_ATTR_PROTOCOL:
			switch native.Uint16(attr.Value) {
			case syscall.IPPROTO_TCP:
				s.Protocol = "TCP"
			case syscall.IPPROTO_UDP:
				s.Protocol = "UDP"
			case syscall.IPPROTO_SCTP:
				s.Protocol = "SCTP"
			default:
				return nil, fmt.Errorf("not support protocol %v", native.Uint16(attr.Value))
			}
		case IPVS_SVC_ATTR_ADDR:
			addr = attr.Value
		case IPVS_SVC_ATTR_PORT:
			s.Port = int(binary.BigEndian.Uint16(attr.Value))
		case IPVS_SVC_ATTR_FWMARK:
			s.FWMark = int(native.Uint32(attr.Value))
		case IPVS_SVC_ATTR_SCHED_NAME:
			s.SchedName = nl.BytesToString(attr.Value)
		case IPVS_SVC_ATTR_FLAGS:
			s.Flags = int(native.Uint32(attr.Value))
		case IPVS_SVC_ATTR_TIMEOUT:
			s.Timeout = int(native.Uint32(attr.Value))
		case IPVS_SVC_ATTR_NETMASK:
			s.Netmask = int(native.Uint32(attr.Value))
		case IPVS_SVC_ATTR_STATS:
			stats, err := assembleStats(attr.Value)
			if err != nil {
				return nil, err
			}
			s.Stats = stats
		}

	}

	switch s.AddressFamily {
	case "IPv4":
		s.Address = (net.IP)(addr[:4]).String()
	case "IPv6":
		s.Address = (net.IP)(addr[:16]).String()
	default:
		return nil, fmt.Errorf("invalid IP address %v", addr)
	}

	return &s, nil
}

func assembleDestinationInterface(attrs []syscall.NetlinkRouteAttr) (*DestinationEntry, error) {

	var d DestinationEntry
	var addr []byte

	for _, attr := range attrs {

		attrType := int(attr.Attr.Type)

		switch attrType {
		case IPVS_DEST_ATTR_PORT:
			d.Port = int(binary.BigEndian.Uint16(attr.Value))
		case IPVS_DEST_ATTR_FWD_METHOD:
			switch native.Uint32(attr.Value) {
			case IP_VS_CONN_F_MASQ:
				d.Method = "NAT"
			case IP_VS_CONN_F_DROUTE:
				d.Method = "DR"
			case IP_VS_CONN_F_TUNNEL:
				d.Method = "TUN"
			default:
				return nil, errors.New("not support method ")
			}
		case IPVS_DEST_ATTR_WEIGHT:
			d.Weight = int(native.Uint16(attr.Value))
		case IPVS_DEST_ATTR_U_THRESH:
			d.UpperThreshold = int(native.Uint32(attr.Value))
		case IPVS_DEST_ATTR_L_THRESH:
			d.LowerThreshold = int(native.Uint32(attr.Value))
		case IPVS_DEST_ATTR_ACTIVE_CONNS:
			d.ActiveConnections = native.Uint32(attr.Value)
		case IPVS_DEST_ATTR_INACT_CONNS:
			d.InActiveConnections = native.Uint32(attr.Value)
		case IPVS_DEST_ATTR_PERSIST_CONNS:
			d.PersistConnections = native.Uint32(attr.Value)
		case IPVS_DEST_ATTR_ADDR_FAMILY:
			switch native.Uint16(attr.Value) {
			case syscall.AF_INET:
				d.AddressFamily = "IPv4"
			case syscall.AF_INET6:
				d.AddressFamily = "IPv6"
			}
		case IPVS_DEST_ATTR_ADDR:
			addr = attr.Value
		case IPVS_SVC_ATTR_STATS:
			stats, err := assembleStats(attr.Value)
			if err != nil {
				return nil, err
			}
			d.Stats = stats
		}
	}

	switch d.AddressFamily {
	case "IPv6":
		d.Address = (net.IP)(addr[:16]).String()
	default:
		d.Address = (net.IP)(addr[:4]).String()
	}

	return &d, nil
}

func assembleStats(msg []byte) (Stats, error) {

	var s Stats

	attrs, err := nl.ParseRouteAttr(msg)
	if err != nil {
		return s, err
	}

	for _, attr := range attrs {
		attrType := int(attr.Attr.Type)
		switch attrType {
		case IPVS_STATS_ATTR_CONNS:
			s.Connections = native.Uint32(attr.Value)
		case IPVS_STATS_ATTR_INPKTS:
			s.PacketsIn = native.Uint32(attr.Value)
		case IPVS_STATS_ATTR_OUTPKTS:
			s.PacketsOut = native.Uint32(attr.Value)
		case IPVS_STATS_ATTR_INBYTES:
			s.BytesIn = native.Uint64(attr.Value)
		case IPVS_STATS_ATTR_OUTBYTES:
			s.BytesOut = native.Uint64(attr.Value)
		case IPVS_STATS_ATTR_CPS:
			s.CPS = native.Uint32(attr.Value)
		case IPVS_STATS_ATTR_INPPS:
			s.PPSIn = native.Uint32(attr.Value)
		case IPVS_STATS_ATTR_OUTPPS:
			s.PPSOut = native.Uint32(attr.Value)
		case IPVS_STATS_ATTR_INBPS:
			s.BPSIn = native.Uint32(attr.Value)
		case IPVS_STATS_ATTR_OUTBPS:
			s.BPSOut = native.Uint32(attr.Value)
		}
	}
	return s, nil
}

type ipvsFlags struct {
	flags uint32
	mask  uint32
}

func (f *ipvsFlags) Serialize() []byte {
	return (*(*[unsafe.Sizeof(*f)]byte)(unsafe.Pointer(f)))[:]
}

func (f *ipvsFlags) Len() int {
	return int(unsafe.Sizeof(*f))
}
