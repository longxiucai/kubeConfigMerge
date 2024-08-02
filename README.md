# 说明
* 如果多个文件内context的名称相同，则使用`文件名-对应的cluster的hash值`作为context名称
* 自动选择最后合并的context作为current-context

# 用法
1、指定多个kubeconfig文件名称合并，中间用逗号`,`分隔，默认输出为./merge.config
```
kubeconfigmerge --file=cluster-1,cluster-2,cluster-3
```
2、指定输出文件名
```
kubeconfigmerge --file=cluster-1,cluster-2,cluster-3 --output=~/.kube/config
```
3、自动读取当前目录下的所有kubeconfig文件并且合并
```
kubeconfigmerge --file .
```