/*
* @Author: BlahGeek
* @Date:   2016-06-13
* @Last Modified by:   BlahGeek
* @Last Modified time: 2016-06-22
 */

package raftsc

import "labrpc"
import "crypto/rand"
import "math/big"
import "sync/atomic"
import "log"
import "os"
import "fmt"

type RaftClient struct {
	servers []*labrpc.ClientEnd
	leader  int

	service_name string
	client_id    int64
	op_id        int64

	logger *log.Logger
}

func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}

func MakeClient(servers []*labrpc.ClientEnd, service_name string) *RaftClient {
	id := nrand()
	ck := &RaftClient{
		servers:      servers,
		client_id:    id,
		service_name: service_name,
		logger:       log.New(os.Stderr, fmt.Sprintf("[RaftClient %v]", id), log.LstdFlags),
	}
	ck.logger.Printf("New clerk inited\n")
	return ck
}

func (ck *RaftClient) DoExec(typ OpType, data interface{}, retry bool) (bool, interface{}) {
	var ok bool

	op := Op{
		Client: ck.client_id,
		Id:     atomic.AddInt64(&ck.op_id, 1),
		Type:   typ,
		Data:   data,
	}

	for {
		var reply OpReply
		ck.logger.Printf("Try executing %v to leader %v\n", op, ck.leader)
		ok = ck.servers[ck.leader].Call(ck.service_name+".Exec", op, &reply)
		ck.logger.Printf("Exec result: %v\n", reply)
		if !ok {
			ck.logger.Printf("RPC fail, retry=%v\n", retry)
			if !retry {
				return false, nil
			}
		} else if reply.Status == STATUS_WRONG_LEADER {
			ck.leader = (ck.leader + 1) % len(ck.servers)
			ck.logger.Printf("Wrong leader, pick next: %v\n", ck.leader)
		} else {
			ck.logger.Printf("OK, reply = %v\n", reply)
			return true, reply.Data
		}
	}
}

func (ck *RaftClient) Exec(typ OpType, data interface{}) interface{} {
	_, ret := ck.DoExec(typ, data, true)
	return ret
}
