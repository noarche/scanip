### Quick Install 

`go install github.com/noarche/scanip/cmd/scanip@latest`


### Manual Install

`go mod init scanip`

`go mod tidy`



`go get github.com/fatih/color`

`go get github.com/malfunkt/iprange`

`go run scanip.go`

`go build scanip.go`
`./scanip`




=============================================================



Usage Examples
🌍 Scan a Single CIDR

go run scanip.go

Then enter:

Enter CIDR(s) or IP range (comma-separated): 192.168.1.0/24
Enter number of threads (default 125): 50
Enter website port (default 80): 8080

🌍 Scan Multiple CIDRs

go run scanip.go

Then enter:

Enter CIDR(s) or IP range (comma-separated): 192.168.1.0/24,10.0.0.1-10.0.0.50
Enter number of threads (default 125): 100
Enter website port (default 80): (Press Enter to use default)

🔎 Enable Verbose Output

go run scanip.go -v

Shows additional output.
📖 Show Help Message

go run scanip.go -h

📁 Where Are the Results Saved?

Results are saved in scanip.results.txt, and the format is:

192.168.1.1,80,My Website,12KB
10.0.0.5,80,Admin Panel,8KB

It updates live while scanning.
✅ Done!

Now you have a functional IP scanner that finds web servers efficiently. 🚀
Need any modifications? Let me know!
