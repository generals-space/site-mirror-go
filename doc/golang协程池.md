# golang协程池

golang并没有原生的协程池, 虽然一个常规的协程池实现起来很简单, 只不过用`for{}`循环创建指定数量的`goroutine`, 各`goroutine`中用`for{}`死循环监听同一个channel对象就可以对这个channel队列协同处理了. 但是实际上一个协程池并没有那么简单.

比如如果一个goroutine出现异常而退出后, 如何补全协程池数量? 不然一个个都退出了, 最后就没了...

再比如如何动态调整协程池数量? 获取正在被占用/空闲的协程数量?

这些问题不得不考虑.
