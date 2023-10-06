## 实现原理

## 使用场景

- 停止信号监听
- 定时任务
- 生产方消费方解耦
- 控制并发的数量

## 阻塞和非阻塞

发送操作，编译器会调用 `runtime.chansend` 函数

```go
func chansend(c * hchan, ep unsafe.Pointer, block bool, callerpc uintptr) bool
```

**阻塞式**

调用 chansend 函数，并且 block = true

**非阻塞式**

调用 chansend 函数，并且 block = false

```go
select {
    case ch <- 10:
    ...
    default:
}
```

**发送**

- 如果 channel 的读等待队列存在接收者 goroutine
    - 将数据直接发送给第一个等待的 goriutine，并且唤醒接收的 goroutine
- 如果 channel 的读等待队列不存在接收者 goroutine
    - 如果循环数组 buf（缓冲区）未满，那么把数据发送到循环数组 buf（缓冲区）的队尾
    - 如果循环数组 buf 已满，那么阻塞发送的流程，将当前 goroutine 加入写等待队列，并且挂起等待唤醒

**接收**

接收操作，编译器会调用 `runtime.chanrecv` 函数

```go
func chanrecv(c * hchan, ep unsafe.Pointer, block bool) (selected, received bool)
```

**阻塞式**

调用 chanrecv 函数，并且 block = true

**非阻塞**

调用 chanrecv 函数，并且 block = false

```go
select {
    case k, v := < ch:
    ...
    default:
}
```



