package mr

import (
	"log"
	"strconv"
	"sync"
	"time"
)
import "net"
import "os"
import "net/rpc"
import "net/http"

type Coordinator struct {
	lock                   sync.Mutex
	fileList               []string
	nMap                   int
	nReduce                int
	availableMapTaskNum    int
	availableReduceTaskNum int
	taskMap                map[string]Task
	availableTasks         chan Task
}

func (c *Coordinator) RPCHandler(args *ApplyForTaskArgs, reply *ApplyForTaskReply) error {
	task, ok := <-c.availableTasks
	if !ok {
		return nil
	}
	log.Printf("Assign task %s_%v to worker %v", task.TaskType, task.Index, args.WorkerId)
	task.WorkerId = args.WorkerId
	task.Deadline = time.Now().Add(10 * time.Second)

	reply.Task = task
	reply.ReduceNum = c.nReduce
	return nil
}

func (c *Coordinator) makeReduceTasks(reduceTaskNum int) {
	for i := 0; i < reduceTaskNum; i++ {
		task := Task{
			Index:    i,
			TaskType: "REDUCE"}
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

func (c *Coordinator) Done() bool {
	ret := false

	// Your code here.
	if c.availableReduceTaskNum == 0 {
		ret = true
	}

	return ret
}

func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{
		lock:                   sync.Mutex{},
		fileList:               files,
		nMap:                   len(files),
		nReduce:                nReduce,
		availableMapTaskNum:    len(files),
		availableReduceTaskNum: nReduce * len(files),
		taskMap:                make(map[string]Task),
		availableTasks:         make(chan Task),
	}
	for i, file := range files {
		task := Task{
			Index:    i,
			TaskType: "MAP",
			FileName: file,
		}
		c.taskMap["MAP_"+strconv.Itoa(i)] = task
		c.availableTasks <- task
	}
	c.server()

	//check whether the Task is under processed
	go func() {
		for {
			time.Sleep(time.Second)

			c.lock.Lock()
			for _, task := range c.taskMap {
				if task.WorkerId != "" && time.Now().After(task.Deadline) {
					log.Printf("Task %v_%v has deadline %v and the process is time-out,"+
						"reassign the task.", task.TaskType, task.Index, task.Deadline)
					task.WorkerId = ""
					c.availableTasks <- task
				}
			}
			c.lock.Unlock()
		}
	}()
	return &c
}
