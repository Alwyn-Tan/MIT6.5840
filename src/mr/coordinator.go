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
	mutex         sync.Mutex
	fileList      []string
	MapTaskNum    int
	ReduceTaskNum int
	task          map[string]Task
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
		mutex:         sync.Mutex{},
		fileList:      files,
		MapTaskNum:    len(files),
		ReduceTaskNum: nReduce * len(files),
		task:          make(map[string]Task),
	}
	for i, file := range files {
		task := Task{
			index:    i,
			taskType: "MAP",
			file:     file,
		}
		c.task["MAP"+strconv.Itoa(i)] = task
	}
	c.server()
	return &c
}
