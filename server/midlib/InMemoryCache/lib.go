package InMemoryCache

import "fmt"

type EtClassType int

const (
	EtagFromMarked EtClassType = 1
	EtagFromProxy  EtClassType = 2
	EtagFromLocal  EtClassType = 3
	EtagUnknown    EtClassType = 4
)

// return where the etag came from
func ClasifyETag(etag string) (etClass EtClassType) {
	switch etag[0] {
	case 'A':
		return EtagFromMarked
	case 'B':
		return EtagFromProxy
	case 'C':
		return EtagFromLocal
	default:
		return EtagUnknown
	}
}

func (et EtClassType) String() string {
	switch et {
	case EtagFromMarked:
		return "EtagFromMarked"
	case EtagFromProxy:
		return "EtagFromProxy"
	case EtagFromLocal:
		return "EtagFromLocal"
	case EtagUnknown:
		return "EtagUnknown"
	default:
		return fmt.Sprintf("-- Unkown Etag* = %d --", int(et))
	}
}
