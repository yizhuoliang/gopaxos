#!  /usr/bin/osascript

tell application "Terminal"
    set window1 to do script "cd gopaxos"
    set window2 to do script "cd gopaxos"
    set window3 to do script "cd gopaxos"
    set window4 to do script "cd gopaxos"
    set window5 to do script "cd gopaxos"
    set window6 to do script "cd gopaxos"
    set window7 to do script "cd gopaxos"
    set window8 to do script "cd gopaxos"
    set window9 to do script "cd gopaxos"
    delay 1
    do script "go mod tidy" in window1
    do script "go mod tidy" in window2
    do script "go mod tidy" in window3
    do script "go mod tidy" in window4
    do script "go mod tidy" in window5
    do script "go mod tidy" in window6
    do script "go mod tidy" in window7
    do script "go mod tidy" in window8
    do script "go mod tidy" in window9
    delay 1
    do script "go run acceptor/main.go 0 0" in window1
    do script "go run acceptor/main.go 1 0" in window2
    do script "go run acceptor/main.go 2 0" in window3
    delay 1
    do script "go run leader/main.go 0 0" in window4
    do script "go run leader/main.go 1 0" in window5
    delay 1
    do script "go run replica/main.go 0 0" in window6
    do script "go run replica/main.go 1 0" in window7
    delay 1
    do script "go run client/main.go 0 0" in window8
    do script "go run client/main.go 1 0" in window9
end tell