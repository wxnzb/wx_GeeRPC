## 1
func numIslands(grid [][]byte) int {
    ans:=0
    m,n:=len(grid),len(grid[0])
    var dfs func(int,int)
    dfs=func(i,j int){
        //0<i<m
        if i<0||i>=m||j<0||j>=n||grid[i][j]!='1'{
            return
        }
        grid[i][j]=2
        dfs(i,j+1)
        dfs(i,j-1)
        dfs(i+1,j)
        dfs(i-1,j)
    }这里为啥用var定义dfs用:=好像不行

- 短声明 := 只用于声明变量并推导类型：:= 适用于局部变量的声明并初始化，但它不能直接用来声明类型为 func 的变量。你不能直接将 dfs := func(i, j int) { ... } 来声明一个函数类型的变量。
必须显式声明函数类型：在 Go 中，函数类型的变量需要先定义它的类型（即 func(int, int)），然后再将匿名函数赋值给这个变量。因此，必须先声明 var dfs func(int, int)。