package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/Heliodex/coputer/bundle"
	"github.com/Heliodex/coputer/wallflower/net"
	"github.com/syncthing/notify"
)

func changed(n *net.Node, path, name string) {
	// fmt.Println("Updating program  ", path)

	b, err := bundle.Bundle(path)
	if err != nil {
		fmt.Println("Failed to bundle program:", err)
		return
	}

	if _, err = net.StoreProgram(n.Pk, name, b); err != nil {
		fmt.Println("Failed to store program:", err)
		return
	}
}

func watchPath(n *net.Node, path string) {
	fmt.Printf("Watching path %s for changes...\n", path)
	name := filepath.Base(path)
	fmt.Printf("Your program will be served on your gateway as http://%s.%s.localhost:2517\n", name, n.Pk.EncodeNoPrefix())

	// Make the channel buffered to ensure no event is dropped. Notify will drop
	// an event if the receiver is not able to keep up the sending pace.
	c := make(chan notify.EventInfo, 1)

	if err := notify.Watch(path, c, notify.All); err != nil {
		fmt.Printf("Failed to watch path %s: %v\n", path, err)
		return
	}
	defer notify.Stop(c)

	// Batch events: call changed() only if no new events arrive within 100ms
	var timer *time.Timer
	timeout := func() <-chan time.Time {
		if timer != nil {
			return timer.C
		}
		// Block forever if timer is nil
		return make(chan time.Time)
	}
	for {
		select {
		case <-c:
			// fmt.Printf("Change detected in %s: %s\n", path, event.Event())
			if timer != nil {
				timer.Stop()
			}
			timer = time.NewTimer(100 * time.Millisecond)
		case <-timeout():
			changed(n, path, name)
			timer = nil
		}
	}
}
