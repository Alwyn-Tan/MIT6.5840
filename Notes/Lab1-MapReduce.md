# How does the sequential MapReduce work
Different languages have different styles to deal with string processing such as parsing, seperating and constructing and so on.
In Go, we use
```	
ff := func(r rune) bool { return !unicode.IsLetter(r) }
words := strings.FieldsFunc(contents, ff)
```
```rune``` in Go is an alias for int32, which can represent any Unicode character and ```unicode.IsLetter(r rune)``` judges whether
the input rune is a letter. </br>
So let us say we have a sentence "This is an apple. And that is also an apple." and the output of the codes above would be ```[This, is, an, apple,
 And, that, is, also, an, apple]```. Then put this slice into []KeyValue and sort it we get out intermediate slice: 
```
[ {"And", "1"},
  {"also", "1"},
  {"an", "1"},
  {"an", "1"},
  {"apple", "1"},
  {"apple", "1"},
  {"is", "1"},
  {"is", "1"},
  {"That", "1"},
  {"This", "1"}  ]
```
In fact, it's ok to directly pass []KeyValue to Reduce(), but that would make Reduce() a little complicated.
To obtain the count of a word, the easier way is to use a bit string. Codes like:
```
	for i < len(intermediate) {
		j := i + 1
		for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
			j++
		}
		var values []string
		for k := i; k < j; k++ {
			values = append(values, intermediate[k].Value)
		}
		output := reducef(intermediate[i].Key, values)
		i = j
	}
```
Still for the intermediate above, the ```values``` would be something like this:
```
"And", ["1"]
"also", ["1"]
"an",["1", "1"]
"apple", ["1", "1"]
...
```
So the final output would be:
```
"And", "1"
"also", "1"
"an", "2"
"apple", "2"
...
```

# How to implement distributed MapReduce?
It is usually a good practice to develop your habits for problem-solving, and my way is asking such questions over and over again, advancing gradually and deeply:
* Can we divide the problem into small parts? And can we analyse the relations among these small parts?
* For each part, what can it do and what information should it know?
* How is the way these parts communicate with others?

It is often not easy to directly figure out the whole picture and procedure for a software, especially for a distributed system, 
but we could try to solve it step by step.
Ok, so for this distributed MapRecuce, let's start by answering some basic questions:

* Yes, this work can be done by one coordinator with some workers. The relation between them are like server-client, many clients ask server for resources, that is exactly the way our coordinator and workers perform
* Coordinators prepare and assign jobs, leadning the work, meanwhile the workers are asking for jobs, doing the jobs and reporting to coordinator continuously.
* MIT courses have told us that what is RPC, that is the way they would communicate in.

## Define the Coordinator
Our coordinator should do such things:
* Prepare the tasks, that means storing the input files, and record the tasks and where the tasks have gone, also the dealine of the tasks
* Control the parallel acceess to tasks, one task should only be accessed by one worker at one moment
* Lead the job stage, when to map, when to reduce and when to finish the job

So, maybe we could try to build a rough coordinator like this:
```
type Coordinator struct {
	lock		sync.Mutex 
	stage		string
	fileList	[]string
	nMap		int
	nReduce		int
	taskMap	 	map[string]Task
}
```

### Task as the resource being transported
You must have noticed that we use ```map[string]Task```, you must have heard something like frame, segment or packet in network, there are the units of data in different layers. Like the network, our coordinator and workers need something to communicate with, that is ```Task```. Workers ask for ```Task``` and coordinator assign ```Task```, to better indentify ```Task``` and know whereit is gone, we define the ```Task```:
```
type Task struct{
	Index		int
	TaskType	string //"Map" or "Reduce" or ""
	FileName	string
	WorkerId 	string
	Deadline 	time.Time
}
```

### Control the stage: Task channel
Should we wait for all workers finish the map tasks then move to reduce tasks, or do map tasks and reduce tasks at the same time? I choose the simple one, which means split the work into map stage and reduce stage, after finishing map stage can we then step into reduce stage. </br>
 To control the stage means we need to know when all map/reduce tasks are all finished, which data structure shuold we use? List? Map? Or...channel? </br>
We often say Go is born for parallel program, and channel is a powerful tool we can use for parallel communicating. It enables us to exchange data between different Goroutines securely. And that is excatly what we want. </br>
So the enhanced coordinator is like:
```
type Coordinator struct {
	lock			sync.Mutex 
	stage			string
	fileList		[]string
	nMap			int
	nReduce			int
	taskMap	 		map[string]Task
	availableTasks	chan Task
}
```

## Communication based on RPC 
One difficulty in this project is how to design RPC systems. And it's quite cool to implement our own protocols.
You　may remember how [TCP][https://en.wikipedia.org/wiki/Transmission_Control_Protocol#Connection_establishment] works.
In a network system, identification is an essential problem. In TCP, we use 'Three-way Handshake' to establish a connection. Double ACKs
ensure that we connect to a trusted address.</br>
What about the distributed Map-Reduce system? We need to verify that the Task was completed by the designated worker, which means worker
should report to coordinator that what Task it has done(ACK), and coordinator then verify that 'yes, what you have done is verified'(ACK):
```
\\in rpc.go
type ApplyForTaskArg struct {
WorkerId      string
LastTaskType  string
LastTaskIndex int
}

\\........

\\in coordinator.go
lastTaskId := generateTaskId(args.LastTaskType, args.LastTaskIndex)
		//check the last task having been finished by the designated worker
		if task, ok := c.taskMap[lastTaskId]; ok && task.WorkerId == args.WorkerId {
			log.Printf("Task %s finished by worker %s", lastTaskId, args.WorkerId)
			...
		}
```





