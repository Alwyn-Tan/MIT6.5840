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



