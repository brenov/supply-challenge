package main

// Imports
import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

// Global constant
const MAX_VALUE int = 5000

/**
 * Product definition
 */
type Product struct {
	id float32
}

/**
 * Counter definition
 */
type Counter struct {
	val    int
	finish bool
	closed bool
	mux    sync.Mutex
}

/**
 * Prints the log with the product ID, start time of production and end time of
 * production.
 */
func (p *Product) printLog(prodCon string, start, end time.Time) {
	fmt.Println("Product", p.id, "processed by ", prodCon, " successfully.", "\n",
		"Time Init: ", fmt.Sprintf("%dH%dm%ds", start.Hour(), start.Minute(), start.Second()), "\n",
		"Time End:  ", fmt.Sprintf("%dH%dm%ds", end.Hour(), end.Minute(), end.Second()), "\n",
		"--------------------------------------------------------------")
}

/**
 * Check if the channel has already being closed.
 * Returns true if the channel has already being closed, false otherwise
 */
func (c *Counter) isClosed() bool {
	c.mux.Lock()
	var bo bool
	// If it's being called for the first time
	if !c.closed {
		c.closed = true
		bo = false
	} else {
		bo = true
	}
	defer c.mux.Unlock()
	return bo
}

/**
 * Check if the counter reached the max value.
 * Returns true if the counter reached the max value, false otherwise.
 */
func (c *Counter) isFinish() bool {
	c.mux.Lock()
	// Check counter
	if c.val < MAX_VALUE-1 {
		c.val++
		c.finish = false
		c.closed = false
	} else if c.val == MAX_VALUE-1 {
		c.val++
		c.closed = false
	} else {
		c.finish = true
	}
	defer c.mux.Unlock()
	return c.finish
}

/**
 * Consume channel items.
 */
func consumer(id string, prodch <-chan Product, wg *sync.WaitGroup) {
	var start time.Time
	var end time.Time
	// Infinite loop
	for {
		start = time.Now()
		// Get the product and a boolean that verifies if the channel is open
		prod, open := <-prodch
		// If the channel is closed
		if !open {
			// Leave the loop
			break
		}
		// Wait
		time.Sleep(100 * time.Millisecond)
		end = time.Now()
		prod.print(id, start, end)
	}
	// Decrement the waitgroup counter
	wg.Done()
}

/**
 * Produce channel items.
 */
func producer(id int, prodch chan Product, wg *sync.WaitGroup, counter *Counter) {
	var start time.Time
	var end time.Time
	count := 1
	// Infinite loop
	for {
		start = time.Now().UTC()
		// Wait
		time.Sleep(100 * time.Millisecond)
		// Create the product
		var prod Product
		//id.count
		prod.id = float32(id) + (float32(count) / 100)
		// Check if the total products hasn't already reached the maximum value
		if !counter.isFinish() {
			// Put the product inside the channel
			prodch <- prod
		} else {
			break
		}
		end = time.Now().UTC()
		prod.print("produtor "+strconv.Itoa(id), start, end)
		count++
	}
	// Check if another productor hasn't close the channel yet
	if !counter.isClosed() {
		// Close the channel
		close(prodch)
	}
	// Decrement the waitgroup counter
	wg.Done()
}

/**
 * Start program execution.
 */
func main() {
	// Create a waitgroup
	var wg sync.WaitGroup
	// Create a counter
	var count Counter
	// Initiate the number of consumers
	var consumers = 5
	// Initiate the number of products
	var producers = 10
	// Reading of arguments for number of consumers and producers, respectively
	if len(os.Args) >= 2 {
		consumers, _ = strconv.Atoi(os.Args[1])
	}
	if len(os.Args) >= 3 {
		producers, _ = strconv.Atoi(os.Args[2])
	}
	// Create a product channel
	cs := make(chan Product, 5000)
	// Inititate the waitgroup with the goroutines number
	// when the count gets to 0 all goroutines blocked are released
	wg.Add(consumers + producers)
	// Starts all consumers routines
	// Needs to be called before the product creation to avoid deadlock
	for i := 0; i < consumers; i++ {
		var id = "consumidor " + strconv.Itoa(i)
		go consumer(id, cs, &wg)
	}
	// Create the producers
	for i := 0; i < producers; i++ {
		go producer(i, cs, &wg, &count)
	}
	// Make the main wait until the wg counter gets to zero
	wg.Wait()
	fmt.Printf("Todos os produtos foram consumidos. \n")
}
