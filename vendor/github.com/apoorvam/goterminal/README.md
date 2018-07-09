# goterminal [![GoDoc](https://godoc.org/github.com/apoorvam/goterminal?status.svg)](https://godoc.org/github.com/apoorvam/goterminal)
A cross-platform Go-library for updating progress in terminal.

## Installation

````
  go get -u github.com/apoorvam/goterminal
````

## Usage

```go
    func main() {
    	// get an instance of writer
    	writer := goterminal.New()

    	for i := 0; i < 100; i = i + 10 {
    		// add your text to writer's buffer
    		fmt.Fprintf(writer, "Downloading (%d/100) bytes...\n", i)
    		// write to terminal
    		writer.Print()
    		time.Sleep(time.Millisecond * 200)

    		// clear the text written by previous write, so that it can be re-written.
    		writer.Clear()
    	}

    	// reset the writer
    	writer.Reset()
    	fmt.Println("Download finished!")
    }
```
## Example 

![output](doc/download_progress.gif)

Another example which uses the [go-colortext](github.com/daviddengcn/go-colortext) library to re-write text along with using colors. Here is output of example:

![output](doc/color_terminal.gif)

Examples can be found [here](https://github.com/apoorvam/goterminal/tree/master/examples).

### More on usage

* Create a `Writer` instance.
* Add your text to writer's buffer and call `Print()` to print text on Terminal. This can be called any number of times.
* Call `Clear()` to move the cursor to position where first `Print()` started so that it can be over-written.
* `Reset()` writer, so it clears its state for next Write.

