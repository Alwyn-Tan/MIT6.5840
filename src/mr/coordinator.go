package mr

import (
	"log"
	"strconv"
	"sync"
)
import "net"
import "os"
import "net/rpc"
import "net/http"

type Coordinator struct {
	mutex                  sync.Mutex
	fileList               []string
	nMap                   int
	nReduce                int
	availableMapTaskNum    int
	availableReduceTaskNum int
	taskMap                map[string]Task
}

func (c *Coordinator) RPCHandler(args *ApplyForTaskArgs, reply *ApplyForTaskReply) error {
	if c.availableMapTaskNum > 0 {
		*reply = ApplyForTaskReply{
			c.taskMap["MAP_"+strconv.Itoa(c.availableMapTaskNum-1)], c.nReduce,
		}
		c.availableReduceTaskNum--

	} else if c.availableReduceTaskNum > 0 {
		*reply = ApplyForTaskReply{
			c.taskMap["REDUCE_"+strconv.Itoa(c.availableReduceTaskNum-1)],
			c.nReduce,
		}
		c.availableReduceTaskNum--
	} else {
		*reply = ApplyForTaskReply{
			Task{
				0, "Done", "",
			}, c.nReduce,
		}
	}
	return nil
}

func (c *Coordinator) makeReduceTasks(reduceTaskNum int) {
	for i := 0; i < reduceTaskNum; i++ {
		task := Task{
			index:    i,
			taskType: "REDUCE"}
		c.taskMap["REDUCE"+strconv.Itoa(i)] = task
	}
}

func (c *Coordinator) server() {
	log.Printf("Coordinator starting server")
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
//
func (c *Coordinator) Done() bool {
	ret := false

	// Your code here.


	return ret
}

func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{
		mutex:                  sync.Mutex{},
		fileList:               files,
		nMap:                   len(files),
		nReduce:                nReduce,
		availableMapTaskNum:    len(files),
		availableReduceTaskNum: nReduce * len(files),
		taskMap:                make(map[string]Task),
	}
	for i, file := range files {
		task := Task{
			index:    i,
			taskType: "MAP",
			fileName: file,
		}
		c.taskMap["MAP"+strconv.Itoa(i)] = task
	}
	c.server()
	return &c
}
