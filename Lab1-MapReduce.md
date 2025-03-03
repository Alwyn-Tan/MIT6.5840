# Notes for implementing Map() and Reduce() 
Different languages have different styles to deal with string processing such as parsing, seperating and constructing and so on.
In Go, we use
```	
ff := func(r rune) bool { return !unicode.IsLetter(r) }
words := strings.FieldsFunc(contents, ff)
```
```rune``` in Go is an alias for int32, which can represent any Unicode character and ```unicode.IsLetter(r rune)``` judges whether
the input rune is a letter. </br>
So let us say we have a sentence "This is an apple. And that is also an apple." and the output of the codes above would be ```[This, is, an, apple,
 And, that, is, also, an, apple]```. Then put this slice into []KeyValue and sort it we get: 
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

# How does the sequential MapReduce work
