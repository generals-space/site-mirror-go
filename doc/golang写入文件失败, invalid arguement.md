# golang写入文件失败, invalid arguement

```
file, _ := os.OpenFile(filePath, os.O_RDWR, os.ModePerm)
_, err = file.Write(fileContent)
```

err不为空, 值为`invalid arguement`

后来发现是打开文件的标识位不正确, 创建文件还应加上`O_CREATE`, 干脆直接使用`os.Create()`函数.