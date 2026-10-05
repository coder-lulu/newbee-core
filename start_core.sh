netstat -antp | grep 910 | grep LISTEN | awk '{print $7}' | awk -F/ '{print $1}' | xargs kill 2>/dev/null
cd rpc
nohup go run core.go -f etc/core.yaml 2>/dev/null &
sleep 5
cd ../api && nohup go run core.go -f etc/core.yaml 2>/dev/null &
