package bgp

import (
	"net"
)

const (
	_ = iota
	typeOpen
	typeUpdate
	typeNotification
	typeKeepalive
	typeRouteRefresh // See RFC 2918

	headerLen = 19
	MaxSize   = 4096 // Maximum size of a BGP message.
	Version   = 4    // Current defined version of BGP.
)

type Message interface {
	pack([]byte) (int, error)
	unpack([]byte) (int, error)
	//	String() string
	Len() int
}

// Header is the fixed-side header for each BGP message. See
// RFC 4271, section 4.1. The marker is omitted.
type Header struct {
	Length uint16
	Type   uint8
}

func newHeader(typ int) *Header { return &Header{0, uint8(typ)} }

type Prefix net.IPNet

// Size returns the length of the mask in bits.
func (p *Prefix) Size() int {
	_, bits := p.Mask.Size()
	return bits
}

// len returns the length of prefix in bytes.
func (p *Prefix) len() int { return 1 + len(p.IP) }

// Path Flags.
const (
	FlagOptional   = 1 << 8
	FlagTransitive = 1 << 7
	FlagPartial    = 1 << 6
	FlagLength     = 1 << 5
)

// Path Codes.
const (
	_ = iota
	Origin
	ASPath
	NextHop
	MultiExitDisc
	LocalPref
	AtomicAggregate
	Aggregator
)

type Path struct {
	Flags uint8
	Code  uint8
	Value []byte
}

func (p *Path) len() int {
	if p.Flags&FlagLength == FlagLength {
		return 2 + 2 + len(p.Value)
	}
	return 2 + 1 + len(p.Value)
}

type Parameter struct {
	Type  uint8
	Value []byte
}

func (p *Parameter) len() int { return 2 + len(p.Value) }

// OPEN holds the information used in the OPEN message format. RFC 4271, Section 4.2.
type OPEN struct {
	*Header
	Version       uint8
	MyAS          uint16
	HoldTime      uint16
	BGPIdentifier net.IP // Must always be a 4 bytes.
	Parameters    []Parameter
}

// NewOPEN returns an initialized OPEN message.
func NewOPEN(MyAS, HoldTime uint16, BGPIdentifier net.IP, Parameters []Parameter) *OPEN {

	return &OPEN{Header: newHeader(typeOpen), Version: Version, MyAS: MyAS,
		HoldTime: HoldTime, BGPIdentifier: BGPIdentifier.To4(), Parameters: Parameters}
}

// Len returns the length of the entire OPEN message.
// When called is also sets the length in m.Length.
func (m *OPEN) Len() int {
	l := 0
	for _, p := range m.Parameters {
		l += p.len()
	}
	return headerLen + 10 + l
}

// UPDATE holds the information used in the UPDATE message format. RFC 4271, section 4.3
type UPDATE struct {
	*Header
	withdrawnRoutesLength uint16 // make implicit
	WithdrawnRoutes       []Prefix
	pathsLength           uint16 // make implicit
	Paths                 []Path
	ReachabilityInfo      []Prefix
}

// NewUPDATE returns an initialized UPDATE message.
func NewUPDATE(WithdrawnRoutes []Prefix, Paths []Path, ReachabilityInfo []Prefix) *UPDATE {

	return &UPDATE{Header: newHeader(typeUpdate),
		WithdrawnRoutes: WithdrawnRoutes, Paths: Paths, ReachabilityInfo: ReachabilityInfo}
}

func (m *UPDATE) Len() int {
	l := 0
	for _, p := range m.WithdrawnRoutes {
		l += p.len()
	}
	for _, p := range m.Paths {
		l += p.len()
	}
	for _, p := range m.ReachabilityInfo {
		l += p.len()
	}

	return headerLen + 4 + l
}

// KEEPALIVE holds only the header and is used for keep alive pings.
type KEEPALIVE struct {
	*Header
}

// NewKEEPALIVE returns an initialized KEEPALIVE message.
func NewKEEPALIVE() *KEEPALIVE { return &KEEPALIVE{Header: newHeader(typeKeepalive)} }
func (m *KEEPALIVE) Len() int  { return headerLen }

// NOTIFICATION holds an error. The TCP connection is closed after sending it.
type NOTIFICATION struct {
	*Header
	ErrorCode    uint8
	ErrorSubcode uint8
	Data         []byte
}

func (m *NOTIFICATION) Len() int { return headerLen + 2 + len(m.Data) }
