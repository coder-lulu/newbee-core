netstat -antp | grep 910 | grep LISTEN | awk '{print $7}' | awk -F/ '{print $1}' | xargs kill 2>/dev/null
