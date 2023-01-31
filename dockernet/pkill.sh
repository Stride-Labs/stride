pkill -f "docker-compose .*stride.* logs" | true
pkill -f "/bin/bash.*create_logs.sh" | true
pkill -f "tail .*.log" | true
