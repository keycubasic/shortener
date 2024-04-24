go tool pprof -http=":9090" -seconds=30 http://localhost:8080/debug/pprof/heap

curl -sK -v http://localhost:8080/debug/pprof/heap > profiles\base.pprof
go tool pprof -http=":9090" -seconds=30 profiles\base.pprof

curl -sK -v http://localhost:8080/debug/pprof/heap > profiles\result.pprof
go tool pprof -http=":9090" -seconds=30 profiles\result.pprof

go tool pprof -top -diff_base=profiles\base.pprof profiles\result.pprof > profiles\pprof.diff