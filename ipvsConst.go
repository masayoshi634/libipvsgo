package libipvs

// IPVS genl commands
const (
	IPVS_CMD_UNSPEC uint8 = iota

	IPVS_CMD_NEW_SERVICE // add service
	IPVS_CMD_SET_SERVICE // modify service
	IPVS_CMD_DEL_SERVICE // delete service
	IPVS_CMD_GET_SERVICE // get info about specific service

	IPVS_CMD_NEW_DEST //add destination
	IPVS_CMD_SET_DEST //modify destination
	IPVS_CMD_DEL_DEST //delete destination
	IPVS_CMD_GET_DEST //get list of all service dests

	IPVS_CMD_NEW_DAEMON //start sync daemon
	IPVS_CMD_DEL_DAEMON //stop sync daemon
	IPVS_CMD_GET_DAEMON //get sync daemon status

	IPVS_CMD_SET_TIMEOUT //set TCP and UDP timeouts
	IPVS_CMD_GET_TIMEOUT //get TCP and UDP timeouts

	IPVS_CMD_SET_INFO //only used in GET_INFO reply
	IPVS_CMD_GET_INFO //get general IPVS info

	IPVS_CMD_ZERO  //zero all counters and stats
	IPVS_CMD_FLUSH //flush services and dests
)

/* Attributes used in the first level of commands
 */
const (
	IPVS_CMD_ATTR_UNSPEC          int = iota
	IPVS_CMD_ATTR_SERVICE             // nested service attribute //
	IPVS_CMD_ATTR_DEST                // nested destination attribute //
	IPVS_CMD_ATTR_DAEMON              // nested sync daemon attribute //
	IPVS_CMD_ATTR_TIMEOUT_TCP         // TCP connection timeout //
	IPVS_CMD_ATTR_TIMEOUT_TCP_FIN     // TCP FIN wait timeout //
	IPVS_CMD_ATTR_TIMEOUT_UDP         //UDP timeout //
)

// Attributes used to describe a service. Used inside nested attribute
// ipvsCmdAttrService
const (
	IPVS_SVC_ATTR_UNSPEC   int = iota
	IPVS_SVC_ATTR_AF           // address family //
	IPVS_SVC_ATTR_PROTOCOL     // virtual service protocol //
	IPVS_SVC_ATTR_ADDR         // virtual service address //
	IPVS_SVC_ATTR_PORT         // virtual service port //
	IPVS_SVC_ATTR_FWMARK       // firewall mark of service //

	IPVS_SVC_ATTR_SCHED_NAME // name of scheduler //
	IPVS_SVC_ATTR_FLAGS      // virtual service flags //
	IPVS_SVC_ATTR_TIMEOUT    // persistent timeout //
	IPVS_SVC_ATTR_NETMASK    // persistent netmask //

	IPVS_SVC_ATTR_STATS // nested attribute for service stats //

	IPVS_SVC_ATTR_PE_NAME // name of ct retriever //

	IPVS_SVC_ATTR_STATS64 // nested attribute for service stats //
)

// Attributes used to describe a destination (real server). Used
// inside nested attribute ipvsCmdAttrDest.
const (
	IPVS_DEST_ATTR_UNSPEC int = iota
	IPVS_DEST_ATTR_ADDR       // real server address
	IPVS_DEST_ATTR_PORT       // real server port

	IPVS_DEST_ATTR_FWD_METHOD // forwarding method
	IPVS_DEST_ATTR_WEIGHT     // destination weight
	IPVS_DEST_ATTR_U_THRESH   // upper threshold
	IPVS_DEST_ATTR_L_THRESH   // lower threshold

	IPVS_DEST_ATTR_ACTIVE_CONNS  // active connections
	IPVS_DEST_ATTR_INACT_CONNS   // inactive connections
	IPVS_DEST_ATTR_PERSIST_CONNS // persistent connections

	IPVS_DEST_ATTR_STATS // nested attribute for dest stats

	IPVS_DEST_ATTR_ADDR_FAMILY // Address family of address

	IPVS_DEST_ATTR_STATS64 //nested attribute for dest stats
)

/*
 * Attributes used to describe service or destination entry statistics
 *
 * Used inside nested attributes IPVS_SVC_ATTR_STATS, IPVS_DEST_ATTR_STATS,
 * IPVS_SVC_ATTR_STATS64 and IPVS_DEST_ATTR_STATS64.

 */
// IPVS Svc Statistics constancs
const (
	IPVS_STATS_ATTR_UNSPEC   int = iota
	IPVS_STATS_ATTR_CONNS        // connections scheduled
	IPVS_STATS_ATTR_INPKTS       // incoming packets
	IPVS_STATS_ATTR_OUTPKTS      // outgoing packets
	IPVS_STATS_ATTR_INBYTES      // incoming bytes
	IPVS_STATS_ATTR_OUTBYTES     // outgoing bytes

	IPVS_STATS_ATTR_CPS    // current connection rate
	IPVS_STATS_ATTR_INPPS  // current in packet rate
	IPVS_STATS_ATTR_OUTPPS // current out packet rate
	IPVS_STATS_ATTR_INBPS  // current in byte rate
	IPVS_STATS_ATTR_OUTBPS // current out byte rate
	IPVS_STATS_ATTR_PAD
)

/*
   IPVS Connection Flags
   Only flags 0..15 are sent to backup server
*/
// Destination forwarding methods
const (
	IP_VS_CONN_F_FWD_MASK   = 0x0007 //mask for the fwd methods
	IP_VS_CONN_F_MASQ       = 0x0000 //masquerading/NAT
	IP_VS_CONN_F_LOCALNODE  = 0x0001 //local node
	IP_VS_CONN_F_TUNNEL     = 0x0002 //tunneling
	IP_VS_CONN_F_DROUTE     = 0x0003 //direct routing
	IP_VS_CONN_F_BYPASS     = 0x0004 //cache bypass
	IP_VS_CONN_F_SYNC       = 0x0020 //entry created by sync
	IP_VS_CONN_F_HASHED     = 0x0040 //hashed entry
	IP_VS_CONN_F_NOOUTPUT   = 0x0080 //no output packets
	IP_VS_CONN_F_INACTIVE   = 0x0100 //not established
	IP_VS_CONN_F_OUT_SEQ    = 0x0200 //must do output seq adjust
	IP_VS_CONN_F_IN_SEQ     = 0x0400 //must do input seq adjust
	IP_VS_CONN_F_SEQ_MASK   = 0x0600 //in/out sequence mask
	IP_VS_CONN_F_NO_CPORT   = 0x0800 //no client port set yet
	IP_VS_CONN_F_TEMPLATE   = 0x1000 //template, not connection
	IP_VS_CONN_F_ONE_PACKET = 0x2000 //forward only one packet
)
