package main

import (
	"flag"
	"fmt"
	"github.com/nutrun/lentil"
	"log"
	"os"
	"strconv"
)

var addr = flag.String("addr", "0.0.0.0:11300", "Beanstalkd address (host:port)")
var listTubes = flag.Bool("list-tubes", false, "List tubes")
var pauseTube = flag.String("pause-tube", "", "Pause tube, requires delay")
var statsTube = flag.String("stats-tube", "", "Get stats for tube")
var stats = flag.Bool("stats", false, "Get queue stats")
var drain = flag.String("drain", "", "Empty tube by deleting all its jobs")
var tube = flag.String("tube", "default", "Use tube")
var put = flag.String("put", "", "Put job body")
var peek = flag.Int("peek", 0, "Peek job by id")
var peekAll = flag.Bool("peek-all", false, "Peek all available job")
var peekTo = flag.Uint64("peek-to", uint64(0), "Peek all available job")
var peekBuried = flag.Bool("peek-buried", false, "Peek first buried job")
var peekDelayed = flag.Bool("peek-delayed", false, "Peek first delayed job")
var kick = flag.Int("kick", 0, "Move n buried or delayed jobs to ready queue")
var del = flag.Int("delete", 0, "Delete job by id")
var delay = flag.Int("delay", 0, "Pause tube or job delay")
var pri = flag.Int("pri", 0, "Job priority")
var ttr = flag.Int("ttr", 360, "Job time to run")

func main() {
	lentil.ReaderSize = 65536
	flag.Parse()
	log.SetFlags(0)
	q, e := lentil.Dial(*addr)
	err(e)
	if *listTubes {
		tubes, e := q.ListTubes()
		err(e)
		for _, tube := range tubes {
			fmt.Println(tube)
		}
		os.Exit(0)
	}
	if *pauseTube != "" {
		if *delay == 0 {
			log.Fatal("Usage: lentil pause-tube=<tube> delay=<seconds>")
		}
		e := q.PauseTube(*pauseTube, *delay)
		err(e)
		os.Exit(0)
	}
	if *statsTube != "" {
		stats, e := q.StatsTube(*statsTube)
		err(e)
		for k, v := range stats {
			fmt.Printf("%s:%s\n", k, v)
		}
		os.Exit(0)
	}
	if *stats {
		stats, e := q.Stats()
		err(e)
		for k, v := range stats {
			fmt.Printf("%s:%s\n", k, v)
		}
		os.Exit(0)
	}
	if *drain != "" {
		_, e := q.Watch(*drain)
		err(e)
		if *drain != "default" {
			_, e = q.Ignore("default")
			err(e)
		}
		for {
			job, e := q.ReserveWithTimeout(0)
			if e != nil {
				if e.Error() == "TIMED_OUT\r\n" {
					break
				}
				log.Fatal(e)
			}
			fmt.Printf("%d:%s\n", job.Id, job.Body)
			e = q.Delete(job.Id)
			err(e)
		}
		os.Exit(0)
	}
	if *put != "" {
		e = q.Use(*tube)
		err(e)
		id, e := q.Put(*pri, *delay, *ttr, []byte(*put))
		err(e)
		fmt.Printf("id:%d\n", id)
		os.Exit(0)
	}
	if *peek != 0 {
		job, e := q.Peek(uint64(*peek))
		err(e)
		fmt.Printf("%d:%s\n", job.Id, job.Body)
		os.Exit(0)
	}
	if *peekTo != 0 {
		peekToFunc(q)
	}
	if *peekAll == true {
		peekAllFunc(q)
	}
	if *peekBuried {
		e := q.Use(*tube)
		err(e)
		job, e := q.PeekBuried()
		err(e)
		fmt.Printf("%d:%s\n", job.Id, job.Body)
		os.Exit(0)
	}
	if *peekDelayed {
		e := q.Use(*tube)
		err(e)
		job, e := q.PeekDelayed()
		err(e)
		fmt.Printf("%d:%s\n", job.Id, job.Body)
		os.Exit(0)
	}
	if *kick != 0 {
		e := q.Use(*tube)
		err(e)
		kicked, e := q.Kick(*kick)
		err(e)
		fmt.Printf("%d\n", kicked)
		os.Exit(0)
	}
	if *del != 0 {
		e := q.Delete(uint64(*del))
		err(e)
		os.Exit(0)
	}

	flag.Usage()
	os.Exit(1)
}

func peekAllFunc(q *lentil.Beanstalkd) {
	fmt.Println("All active jobs listed bellow:")
	stats, e := q.StatsTube(*tube)
	err(e)
	// fmt.Println(stats["total-jobs"])
	jobAccount, e := strconv.ParseUint(stats["total-jobs"], 10, 64)
	if e != nil {
		fmt.Println(e)
		os.Exit(1)
	}
	var activeJobAccount uint64
	for i := jobAccount; i > 0; i-- {
		job, e := q.Peek(i)
		if e != nil {
			if e.Error() == "NOT_FOUND\r\n" {
				continue
			}

			fmt.Println(e)
			os.Exit(1)
		}
		activeJobAccount += 1
		fmt.Printf("%d => %s\n", job.Id, job.Body)
	}
	fmt.Println("Count of all active jobs: ", activeJobAccount)
	os.Exit(0)
}

func peekToFunc(q *lentil.Beanstalkd) {
	fmt.Println("All active jobs listed bellow:")
	stats, e := q.StatsTube(*tube)
	err(e)
	// fmt.Println(stats["total-jobs"])
	jobAccount, e := strconv.ParseUint(stats["total-jobs"], 10, 64)
	if e != nil {
		fmt.Println(e)
		os.Exit(1)
	}
	var activeJobAccount uint64
	for i := jobAccount; i > *peekTo; i-- {
		job, e := q.Peek(i)
		if e != nil {
			if e.Error() == "NOT_FOUND\r\n" {
				continue
			}

			fmt.Println(e)
			os.Exit(1)
		}
		activeJobAccount += 1
		fmt.Printf("%d => %s\n", job.Id, job.Body)
	}
	fmt.Println("Count of all active jobs: ", activeJobAccount)
	os.Exit(0)
}

func err(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
