package did

type (
	Method string
)

const (
	KeyMethod  Method = "key"
	PeerMethod Method = "peer"
	PKHMethod  Method = "pkh"
	WebMethod  Method = "web"
)

func (m Method) String() string {
	return string(m)
}

// DID represents functionality common to all DIDs
type DID interface {
	// IsValid checks if the DID is compliant with its methods definition
	IsValid() bool
	// ToString Returns the string representation of the DID identifier (e.g. did:example:abcd)
	ToString() string
	// Suffix provides the value of the DID without the method prefix
	Suffix() (string, error)
	// Method provides the method for the DID
	Method() Method
}
