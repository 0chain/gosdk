### Introduction

Demo to try out winsdk

usage:
```
# first compile winsdk from the root of the gosdk project
make build-windows

# copy the generated DLLs into the demo folder
cp winsdk/zcn.windows.* winsdk/demo/

# run demo
cd winsdk/demo
go run main.go
```
