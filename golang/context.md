<!-- date: 2023-09-10 -->

# Context

## 长话短说

context 用来解决 goroutine 之间退出通知、元数据传递的功能。

## 使用 context 的几点建议

1. 不要将 Context 塞到结构体里。直接将 Context 类型作为函数的第一参数，而且一般都命名为 ctx。
2. 不要向函数传入一个 nil 的 context，如果你实在不知道传什么，标准库给你准备好了一个 context：todo。
3. 不要把本应该作为函数参数的类型塞到 context 中，context 存储的应该是一些共同的数据。例如：登陆的 session、cookie 等。
4. 同一个 context 可能会被传递到多个 goroutine，别担心，context 是并发安全的。

## 源码分析

context 是一个接口，官方提供了 `emptyCtx、cancelCtx、timerCtx、valueCtx` 的实现。

```Go
func Context interface {
    Deadline() (deadline time.Time, ok bool)
    Done() <-chan struct{}
    Err() error
    Value(key any) any
}
```

`emptyCtx` 实现了一组空的 `Deadline、Done、Err、Value` 方法。其中 `context.TODO 和 context.Background` 返回的就是 `*emptyCtx`。

```Go
type emptyCtx int

func (*emptyCtx) Deadline() (deadline time.Time, ok bool) {
	return
}

func (*emptyCtx) Done() <-chan struct{} {
	return nil
}

func (*emptyCtx) Err() error {
	return nil
}

func (*emptyCtx) Value(key any) any {
	return nil
}

var (
	background = new(emptyCtx)
	todo       = new(emptyCtx)
)

func Background() Context {
	return background
}

func TODO() Context {
	return todo
}
```

`cancelCtx` 是一种可以取消的 Context. 它继承了 `context.Context`, `done` 用于获取 Context 的取消通知，`children` 用于存储当前节点为根节点的所有可取消的 Context，`err` 用于存储取消时指定的错误信息，`mu` 用来保护 `cancelCtx` 属性的锁。

```Go
type cancelCtx struct {
	Context
	mu       sync.Mutex
	done     atomic.Value
	children map[canceler]struct{}
	err      error
	cause    error
}

// 从 cancelCtx.done 中读取或者存入一个 chan, 并将这个 chan 以只读的方式返回
func (c *cancelCtx) Done() <-chan struct{} {
	d := c.done.Load()
	if d != nil {
		return d.(chan struct{})
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	d = c.done.Load()
	if d == nil {
		d = make(chan struct{})
		c.done.Store(d)
	}
	return d.(chan struct{})
}

// cancel 方法主要做了2件事
// 1. 关闭 cancelCtx.done chan.
// 2. 遍历 cancelCtx.children, 调用 child.cancel 方法
func (c *cancelCtx) cancel(removeFromParent bool, err, cause error) {
    // ...
	c.mu.Lock()
	if c.err != nil {
		c.mu.Unlock()
		return // already canceled
	}
	c.err = err
	c.cause = cause
	d, _ := c.done.Load().(chan struct{})
	if d == nil {
		c.done.Store(closedchan)
	} else {
		close(d)
	}
	for child := range c.children {
		// NOTE: acquiring the child's lock while holding parent's lock.
		child.cancel(false, err, cause)
	}
	c.children = nil
	c.mu.Unlock()
    // ...
}
```

`context.WithCancel` 包装了 `canceCtx` 并提供一个取消函数，调用这个函数可以 Cancel 对应的 Context.

```Go
func WithCancel(parent Context) (ctx Context, cancel CancelFunc) {
	c := withCancel(parent)
	return c, func() { c.cancel(true, Canceled, nil) }
}
```

`timerCtx` 继承了 `canceCtx`，封装了一个定时器和一个截止时间。可以根据需要主动取消，也可以达到 deadline 时通过 timer 来触发取消动作。

```Go
type timerCtx struct {
	*cancelCtx
	timer *time.Timer
	deadline time.Time
}

func (c *timerCtx) cancel(removeFromParent bool, err, cause error) {
	c.cancelCtx.cancel(false, err, cause)
	if removeFromParent {
		// Remove this timerCtx from its parent cancelCtx's children.
		removeChild(c.cancelCtx.Context, c)
	}
	c.mu.Lock()
	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}
	c.mu.Unlock()
}
```

`context.WithTimeout` 和 `context.WithDeadline` 包装了 `timerCtx`, 区别是 `context.WithDeadline` 需要指定一个时间点，而 `context.WithTimeout` 接收一个时间段。

```Go
func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc) {
	return WithDeadline(parent, time.Now().Add(timeout))
}

func WithDeadline(parent Context, d time.Time) (Context, CancelFunc) {
    // ...
	c := &timerCtx{
		cancelCtx: newCancelCtx(parent),
		deadline:  d,
	}
	// ...
	return c, func() { c.cancel(true, Canceled, nil) }
}
```

`valueCtx` 用来存储键值对数据。

```Go
type valueCtx struct {
	Context
	key, val any
}

func (c *valueCtx) Value(key any) any {
	if c.key == key {
		return c.val
	}
	return value(c.Context, key)
}
```

`context.WithValue` 包装了 `valueCtx`，可以给 Context 添加一个键值对信息。

```Go
func WithValue(parent Context, key, val any) Context {
	if !reflectlite.TypeOf(key).Comparable() {
		panic("key is not comparable")
	}
	return &valueCtx{parent, key, val}
}
```

## 示例

### 取消 goroutine

```Go
func WithCancel(parent Context) (ctx Context, cancel CancelFunc)
func WithDeadline(parent Context, deadline time.Time) (Context, CancelFunc)
func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc)
```

```Go
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	fn(ctx)
	fmt.Println(ctx.Err())
}

func fn(ctx context.Context) {
	go func() {
		for {
			fmt.Println("hello world")
			time.Sleep(time.Second * 1)
		}
	}()
	<-ctx.Done()
	fmt.Println("fn done")
}
```

### WithValue

```Go
func main() {
    ctx := context.WithValue(context.Background(), traceId, "app-001")
    process(ctx)
}

var traceId = 0

type traceIdType int

func process(ctx context.Context) {
    traceId, ok := ctx.Value(traceId).(string)
    if ok {
        fmt.Printf("process over. trace_id=%v\n", traceId)
    } else {
        fmt.Println("process over. on trace_id")
    }
}
```

