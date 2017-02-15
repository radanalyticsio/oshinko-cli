package clusters

const NoCodeAvailable = 0
const ClusterConfigCode = 100
const ClientOperationCode = 101
const ClusterIncompleteCode = 102
const NoSuchClusterCode = 103
const ComponentExistsCode = 104

type ClusterError struct {
	Msg  string
	Code int
}

func (e ClusterError) Error() string {
	return e.Msg
}

func NewClusterError(msg string, code int) ClusterError {
	return ClusterError{Msg: msg, Code: code}
}

func ErrorCode(err error) int {
	ce, ok := err.(ClusterError)
	if ok {
		return ce.Code
	}
	return NoCodeAvailable
}
