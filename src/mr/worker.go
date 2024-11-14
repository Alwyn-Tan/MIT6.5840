package mr

import (
	"fmt"
	"io"
	"os"
	"strconv"
)
import "log"
import "net/rpc"
import "hash/fnv"

type KeyValue struct {
	Key   string
	Value string
}

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {
	workerId := os.Getpid()
	log.Printf("Worker %s start\n", workerId)
	for {
		args := ApplyForTaskArgs{
			WorkerId: workerId,
		}
		reply := ApplyForTaskReply{}
		call("Coordinator.RPCHanlder", &args, &reply)
		if reply.Task.taskType == "MAP" {
			callMapFunc(mapf, strconv.Itoa(workerId), reply)
		} else if reply.Task.taskType == "REDUCE" {
			callReduceFunc()
		} else {
			log.Println("Task Done")
			break
		}

	}
}

func callMapFunc(mapf func(string, string) []KeyValue, workerId string, reply ApplyForTaskReply) {
	file, err := os.Open(reply.Task.fileName)
	if err != nil {
		log.Fatalf("cannot open %v", reply.Task.fileName)
	}
	content, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("cannot read %v", reply.Task.fileName)
	}
	file.Close()
	kva := mapf(reply.Task.fileName, string(content))
	kvaHashMap := make(map[int][]KeyValue)
	for _, kv := range kva {
		hashKey := ihash(kv.Key) % reply.ReduceNum
		kvaHashMap[hashKey] = append(kvaHashMap[hashKey], kv)
	}
	for i := 0; i < reply.ReduceNum; i++ {
		oname := "immediate-" + workerId + "-" + strconv.Itoa(i)
		ofile, _ := os.Create(oname)
		for _, kv := range kvaHashMap[i] {
			fmt.Fprintf(ofile, "%v %v\n", kv.Key, kv.Value)
		}
		ofile.Close()
	}
}
func callReduceFunc() {}
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
