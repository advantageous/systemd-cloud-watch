package main

import "github.com/RichardHightower/systemd-cloud-watch/cloud-watch"

func main() {

	journal, err := cloud_watch.NewJournal()
	if err !=nil {

	}
	defer journal.Close()
}
