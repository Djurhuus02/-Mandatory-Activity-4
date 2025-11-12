package main
import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	ra "example.com/ra/internal/ra"
)
func main(){
	var (
		id = flag.String("id","","node id (e.g., n1)")
		addr = flag.String("addr","127.0.0.1:5001","listen address")
		peersCSV = flag.String("peers","","comma separated peer addresses incl. self")
		logdir = flag.String("logdir","logs","directory for log files")
		interval = flag.Duration("interval",5*time.Second,"workload tick interval")
		csHold = flag.Duration("cs_hold",3*time.Second,"time spent inside critical section")
		every = flag.Int("every",6,"request CS every Nth tick")
	)
	flag.Parse()
	if *id==""{ fmt.Println("missing --id"); os.Exit(2) }
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM); defer cancel()
	if err := ra.Run(ctx, *id, *addr, *peersCSV, *logdir, *interval, *csHold, *every); err != nil { fmt.Fprintln(os.Stderr, "error:", err); os.Exit(1) }
}
