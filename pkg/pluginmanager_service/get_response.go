package pluginmanager_service

import (
	pb "github.com/turbot/steampipe/pkg/pluginmanager_service/grpc/proto"
	"sync"
)

// getResponse wraps pb.GetResponse, implementing locking or map access to allow concurrent usage
type getResponse struct {
	*pb.GetResponse

	failureLock  sync.Mutex
	reattachLock sync.Mutex
}

func newGetResponse() *getResponse {
	return &getResponse{
		GetResponse: &pb.GetResponse{
			ReattachMap: make(map[string]*pb.ReattachConfig),
			FailureMap:  make(map[string]string),
		},
	}
}

func (r *getResponse) AddFailure(instance string, s string) {
	r.failureLock.Lock()
	defer r.failureLock.Unlock()
	r.FailureMap[instance] = s
}

func (r *getResponse) AddReattach(c string, reattach *pb.ReattachConfig) {
	r.reattachLock.Lock()
	defer r.reattachLock.Unlock()
	r.ReattachMap[c] = reattach
}
