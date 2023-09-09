# defer 实现原理

**归纳总结**

- defer 定义的延迟函数的参数在 defer 语句出时就已经确定下来了
- defer 定义顺序与实际执行顺序相反
- return 不是原子级操作的，执行过程是: 给返回值赋值 —> 执行 defer —> 执行 return 指令

**使用场景**

释放锁、捕获野生 goroutine 的 panic

**实现原理**

编译器会将 defer 函数直接插入到函数的尾部。把 defer 函数在当前函数内展开直接调用。

源码：

```Go
func A(i int) {
    defer A1(i, 2*i)
    if (i > 1) {
        defer A2("hello", "world")
    }
    // code to do something
    return
}

func A1(i int, j int) {
    // code to do something
}

func A2(m, n string) {
    // code to do something
}
```

编译后：

```Go
func A(i int) {
    // code to do something
    if (i > 1) {
        A2("hello", "world")
    }
    A1(i, 2*i)
    return
}
```

**示例**

- 函数退出前，按照先进后出的顺序执行
- panic 后的 defer 函数不会执行
  
  ```Go
  func main() {
    defer fmt.Println("panice before")
    panic("发生 panic")
    defer func() {
      fmt.Println("panice after")
    }()
  }

  // panic before
  // panic: 发生 panic
  ```

- panic 没有被 recover 时，抛出的 panic 到当前 goroutine 最上层程序直接异常终止
  
  ```Go
  func main() {
    defer func() {
      fmt.Println("c")
    }()
    f()
    fmt.Println("继续执行")
  }

  func f() {
    defer func() {
      fmt.Println("b")
    }()
    panic("a")
  }

  // b
  // c
  // panic: a
  ```

- panic 有被 recover 时，当前 goroutine 最上层函数正常执行
  
  ```Go
  func main() {
    defer func() {
      fmt.Println("c")
    }()
    f()
    fmt.Println("继续执行")
  }

  func f() {
    defer func() {
        if err := recover(); err != nil {
            fmt.Println("recover:", err)
        }
        fmt.Println("b")
    }
    panic("a")
  }

  // recover: a
  // b
  // 继续执行
  // c
  ```

**更多示例**

```Go
func main() {
  var a = 1
  defer fmt.Println(a)
  
  a = 2
  return
}

// 1
```

> 延迟函数 fmt.Println(a) 的参数在 defer 语句出现的时候就已经确定下来了，所以不管后面如何修改 a 变量，都不会影响延迟函数。

```Go
func main() {
 deferTest()
}

func deferTest() {
 var arr = [3]int{1, 2, 3}
 defer printTest(&arr)
 
 arr[0] = 4
 return
}

func printTest(array *[3]int) {
 for i := range array {
  fmt.Println(array[i])
 }
}

// 4 2 3
```

> 延迟函数 printTest() 的参数在 defer 语句出现的时候就已经确定下来了，即为数组的地址，延迟函数执行的时机是在 return 语句之前，所以对数组的最终修改的值会被打印出来。

```Go
func main() {
 res := deferTest()
 fmt.Println(res)
}

func deferTest () (result int) {
  i := 1
  
  defer func() {
    result++
  }()
  
  return i
}

// 2
```

> 函数的 return 语句并不是原子级的，实际的执行过程为为设置返回值—>ret，defer 语句是在返回前执行，所以返回过程是：「设置返回值—>执行defer—>ret」。所以 return 语句先把 result 设置成 i 的值（1），defer 语句中又把 result 递增 1 ，所以最终返回值为 2 。

```Go
func deferTest () (result int) {
  i := 1
  
  defer func() {
    result++
  }()
  
  return i
}

// 执行过程：
// 1. 设置返回值 result = 1
// 2. 执行defer语句，result++
// 3. return
```

```Go
func test() int {
  var i int
  defer func() {
    i++
  }()
  
  return 1
}

// 执行过程：
// 1. 设置返回值 result = 1
// 2. 执行 defer 语句，i++
// 3. return
```

```Go

func test() int {
  var i int
  defer func() {
    i++
  }()
  
  return i
}

// 执行过程：
// 1. 设置返回值 aaa = 0
// 2. 执行 defer 语句，i++
// 3. return
```

> 对于匿名返回值来说，我们可以假定仍然有一个变量用来存储返回值，例如假定返回值变量为 ”aaa”

```Go
func main() {
 res := test()
 fmt.Println(res) // 1
}
func test() (i int) {
 defer func() {
  i++
 }()
 return 0
}

// 执行过程：
// 1. 设置返回值 i = 0
// 2. 执行 defer 语句，i++
// 3. return
```