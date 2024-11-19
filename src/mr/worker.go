package mr

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)
import "log"
import "net/rpc"
import "hash/fnv"

type KeyValue struct {
	Key   string
	Value string
}

type ByKey []KeyValue

func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {
	workerId := strconv.Itoa(os.Getpid())
	var lastTaskType string
	var lastTaskIndex int
	log.Printf("Worker %s start\n", workerId)
	for {
		args := ApplyForTaskArgs{
			WorkerId:      workerId,
			LastTaskType:  lastTaskType,
			LastTaskIndex: lastTaskIndex,
		}
		reply := ApplyForTaskReply{}
		call("Coordinator.RPCHandler", &args, &reply)
		if reply.Task.TaskType == "MAP" {
			callMapFunc(mapf, workerId, reply)
		} else if reply.Task.TaskType == "REDUCE" {
			callReduceFunc(reducef, workerId, reply)
		} else {
			log.Println("Task Done")
			break
		}
		lastTaskType = reply.Task.TaskType
		lastTaskIndex = reply.Task.Index
		log.Printf("Worker %s finish task %s \n", workerId, reply.Task.TaskType+"_"+strconv.Itoa(reply.Task.Index))
	}
}

func callMapFunc(mapf func(string, string) []KeyValue, workerId string, reply ApplyForTaskReply) {
	file, err := os.Open(reply.Task.FileName)
	if err != nil {
		log.Fatalf("cannot open %v", reply.Task.FileName)
	}
	content, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("cannot read %v", reply.Task.FileName)
	}
	file.Close()
	kva := mapf(reply.Task.FileName, string(content))
	kvaHashMap := make(map[int][]KeyValue)
	for _, kv := range kva {
		hashKey := ihash(kv.Key) % reply.ReduceNum
		kvaHashMap[hashKey] = append(kvaHashMap[hashKey], kv)
	}
	for i := 0; i < reply.ReduceNum; i++ {
		oname := tempImmediateFileName(workerId, reply.Task.Index, i)
		ofile, _ := os.Create(oname)
		for _, kv := range kvaHashMap[i] {
			fmt.Fprintf(ofile, "%v %v\n", kv.Key, kv.Value)
		}
		ofile.Close()
	}
}
func callReduceFunc(reducef func(string, []string) string, workerId string, reply ApplyForTaskReply) {
	var lines []string
	for mapIndex := 0; mapIndex < reply.MapNum; mapIndex++ {
		finalMapFile := finalImmediateFileName(mapIndex, reply.Task.Index)

		file, err := os.Open(finalMapFile)
		if err != nil {
			log.Fatalf("cannot open %v", finalMapFile)
		}

		content, err := io.ReadAll(file)
		if err != nil {
			log.Fatalf("cannot read %v", finalMapFile)
		}

		file.Close()
		lines = append(lines, strings.Split(string(content), "\n")...)
	}

	var kva []KeyValue
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, " ")
		kva = append(kva, KeyValue{parts[0], parts[1]})
	}

	sort.Sort(ByKey(kva))

	ofile, _ := os.Create(tempOutputFileName(workerId, reply.Task.Index))

	i := 0
	for i < len(kva) {
		j := i + 1
		for j < len(kva) && kva[j].Key == kva[i].Key {
			j++
		}
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, kva[k].Value)
		}
		output := reducef(kva[i].Key, values)

		// this is the correct format for each line of Reduce output.
		fmt.Fprintf(ofile, "%v %v\n", kva[i].Key, output)

		i = j
	}

	ofile.Close()
}
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
