<!-- date: 2021-08-19 -->

# Slice 是什么结构？

Slice 是一个引用类型，它是由一个指向底层数组的指针、长度（Length）和容量（Capacity）组成的结构。

**指针**：切片包含一个指针，指向底层数组的起始位置。这个指针表示切片的第一个元素。

**长度**：切片的长度是指切片中当前元素的数量。它表示切片的有效数据部分的大小。

**容量**：切片的容量是指切片底层数组中可以容纳的元素数量。它通常大于或等于切片的长度。

# 初始化 Slice

- `var ints []int`

   ints 变量 `data = nil`、`len = 0`、`cap = 0`。

- make `var ints = make([]int, 2, 5)`

   开辟一个容量为 `cap = 5` 并且初始长度为 `len = 2` 的数组，data 指针指向该数组。
  
   ```Go
   func main() {
     ints := make([]int, 2, 5)
     fmt.Println(cap(ints)) // 5
     fmt.Println(len(ints)) // 2
     fmt.Println(ints) // 0,0
   }
   ```

- new `var ints = new([]int)`

   new 关键字创建一个 Slice 和 `var ints []int` 一样，不会初始化该 Slice。

# Slice 和底层数组的关系

```Go
func main() {
  // 底层数组 arr
  arr := [10]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
  // Slice 1
  s1 := arr[1:5]
  // Slice 2
  s2 := arr[7:]
  
  fmt.Println(cap(s1), len(s1), s1) // cap = 9, len = 4, s1 = [1,2,3,4]
  fmt.Println(cap(s2), len(s2), s2)	// cap = 3, lem = 3, s2 = [7,8,9]
  
  // 修改底层数组 arr 的值
  arr[1] = 11
  
  fmt.Println(s1) // s1 = [11,2,3,4]
  
  // Slice 1 新增值
  s1 = append(s1, 55, 66, 77)
  
  fmt.Println(s1) // s1 = [11,2,3,4,55,66,77]
  fmt.Println(s2) // s2 = [77,8,9]
  fmt.Println(arr) // arr = [0,11,2,3,4,55,66,77,8,9]
  
  // 扩容 Slice 2
  s2 = append(s2, 10, 11, 12)
  
  // 修改 Slice 2 底层数组的值
  s2[0] = 777
  
  fmt.Println(s2) // s2 = [777,8,9,10,11,12]
  // Slice 2 发生了扩容，s2 data 跟底层数组 arr 脱离了关系，从而指向一个全新的底层数组
  fmt.Println(arr) // arr = [0,11,2,3,4,55,66,77,8,9]
  fmt.Println(s1) // s1 = [11,2,3,4,55,66,77]
}
```

**总结：** 切片是底层数组的引用，当切片发生扩容时，它会分配一个新的底层数组，并将数据复制到新数组中。修改切片会影响底层数组，但扩容后的切片会与原始底层数组脱离关系。

### Slice 的扩容规则

1. 预估扩容后的容量 newCap

   预估规则：如果 oldCap * 2 < cap 那么 newCap = cap；否则如果 oldLen < 1024 那么 newCap  = oldCap *2（翻倍）；否则如果 oldLen >= 1025 那么 newCap = oldCap * 1.25（扩容 1/4）。

   ```Go
   ints := []int{1, 2}	// cap = 2
   ints = append(ints, 3, 4, 5)	// 2*2 < 5 所以 newCap = 5
   
   ints2 := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}	// cap = 9
   ints2 = append(ints2, 1, 2) // 9*2 > 11, 9 < 1024 所以 newCap = 9 * 2
  
   ints3 := []int{1, 2, 3, ..., 1025}	// cap = 1025
   ints3 = append(ints2, 1026, 1027)	// 1025 >= 1025 所以预估 newCap = 1025 * 1.25
   ```

2. 计算 newCap 个元素需要多大内存

   所需内存 = 预估容量 * 元素类型大小

3. 将预估的内存匹配到合适的内存规格

   Go 内存管理模块管理着不同规格的内存，申请内存时内存管理模块会匹配到最够大且最接近的规格。
   
4. 计算扩容后的容量

   newCap = 内存规格 / 元素所占的内存空间（byte）

```Go
ints := []int{1, 2}
ints = append(ints, 3, 4, 5)

// 1. 预估扩容后的容量 newCap: 5
// 2. 计算 newCap 个元素需要多大内存: 5*8 = 40 byte
// 3. 将预估的内存匹配到合适的内存规格: 48（64位 OS 的内存规格 8 16 32 48 ... byte）
// 4. 计算扩容后的容量: 48/8 = 6
```
