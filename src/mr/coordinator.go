package mr

import (
	"log"
	"math"
	"strconv"
	"sync"
	"time"
)
import "net"
import "os"
import "net/rpc"
import "net/http"

type Coordinator struct {
	lock           sync.Mutex
	stage          string
	fileList       []string
	nMap           int
	nReduce        int
	taskMap        map[string]Task
	availableTasks chan Task
}

func (c *Coordinator) RPCHandler(args *ApplyForTaskArgs, reply *ApplyForTaskReply) error {
	//last task finished handling
	if args.LastTaskType != "" {
		c.lock.Lock()
		lastTaskId := generateTaskId(args.LastTaskType, args.LastTaskIndex)
		//check the last task having been finished
		if task, ok := c.taskMap[lastTaskId]; ok && task.WorkerId == args.WorkerId {
			log.Printf("Task %s finished by worker %s", lastTaskId, args.WorkerId)
			if args.LastTaskType == "MAP" {
				for reduceIndex := 0; reduceIndex < c.nReduce; reduceIndex++ {
					err := os.Rename(tempImmediateFileName(args.WorkerId, args.LastTaskIndex, reduceIndex),
						finalImmediateFileName(args.LastTaskIndex, reduceIndex))
					if err != nil {
						log.Fatalf("Failed to out put final map results on %s", tempImmediateFileName(args.WorkerId, args.LastTaskIndex, reduceIndex))
						return err
					}
				}
			} else if args.LastTaskType == "REDUCE" {
				//for reduce task, LastTaskIndex is exatly the reduce index
				err := os.Rename(tempOutputFileName(args.WorkerId, args.LastTaskIndex),
					finalOutputFileName(args.LastTaskIndex))
				if err != nil {
					log.Fatalf("Failed to out put final reduce results %s", tempOutputFileName(args.WorkerId, args.LastTaskIndex))
					return err
				}
			}
			delete(c.taskMap, lastTaskId)
			if len(c.taskMap) == 0 {
				c.changeStage()
			}
		}
		c.lock.Unlock()
	}

	task, ok := <-c.availableTasks
	if !ok {
		return nil
	}
	log.Printf("Assign task %s_%v to worker %v", task.TaskType, task.Index, args.WorkerId)
	task.WorkerId = args.WorkerId
	task.Deadline = time.Now().Add(10 * time.Second)
	c.taskMap[generateTaskId(task.TaskType, task.Index)] = task
	reply.Task = task
	reply.MapNum = c.nMap
	reply.ReduceNum = c.nReduce
	return nil
}

func (c *Coordinator) changeStage() {
	if c.stage == "MAP" {
		log.Printf("MAP stage is changed to REDUCE stage")
		c.stage = "REDUCE"
		c.makeReduceTasks(c.nReduce)
	} else if c.stage == "REDUCE" {
		log.Printf("REDUCE stage is changed to DONE stage")
		close(c.availableTasks)
		c.stage = "DONE"
	}
}

func (c *Coordinator) makeReduceTasks(reduceTaskNum int) {
	for i := 0; i < reduceTaskNum; i++ {
		task := Task{
			Index:    i,
			TaskType: "REDUCE"}
		c.taskMap["REDUCE_"+strconv.Itoa(i)] = task
		c.availableTasks <- task
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
	return c.stage == "DONE"
}

func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{
		lock:           sync.Mutex{},
		stage:          "MAP",
		fileList:       files,
		nMap:           len(files),
		nReduce:        nReduce,
		taskMap:        make(map[string]Task),
		availableTasks: make(chan Task, int(math.Max(float64(len(files)), float64(nReduce)))),
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
