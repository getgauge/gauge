go-colortext package
====================

This is a package to change the color of the text and background in the console, working both in windows and other systems.

Under windows, the console APIs are used, and otherwise ANSI text is used.

Docs: http://godoc.org/github.com/daviddengcn/go-colortext ([packages that import ct](http://go-search.org/view?id=github.com%2fdaviddengcn%2fgo-colortext))

Usage:
```go
ChangeColor(Red, true, White, false)
fmt.Println(...)
ChangeColor(Green, false, None, false)
fmt.Println(...)
ResetColor()
```
