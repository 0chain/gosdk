package sdk

import "sync"

//Provide interface similar to io.Reader
//Define errors in this file temporarily

type streamDownload struct {
	AllocationID             string
	Blobbers                 []interface{}
	DataShards, ParityShards int
	//Offset Where to start to read from
	Offset int64
	//Encrypted Is file encrypted before uploading
	Encrypted bool
	//Retry Set this value to retry some failed requests due to too_many_requests, context_cancelled, timeout, etc. errors
	Retry int
}

type dataStatus struct {
	err  error
	data []byte
	n    int //Check if we can partially get data
}

type counter struct {
	succeeded int
	failed    int
	mu        sync.Mutex
}

func Read(p []byte) (n int, err error) {
	return
}

func getData() {
	wg := sync.WaitGroup{}
	c := new(counter)

	for i := 0; i < 10; i++ {
		wg.Add(1)

		go downloadDataFromBlobber(&wg, c)
	}
	wg.Wait()

}
func downloadDataFromBlobber(wg *sync.WaitGroup, c *counter) (data []byte, err error) {
	defer wg.Done()

	//Handle too_many_requests, context_cancelled, timeout, etc errors in this function
	c.mu.Lock()
	/*
		if failed{
			c.failed++
		}else{
			c.succeeded++
		}
	*/
	defer c.mu.Unlock()
	return
}
