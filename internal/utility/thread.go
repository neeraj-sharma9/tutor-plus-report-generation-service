package utility

func CreateSemaphore(size int) chan int {
	sem := make(chan int, size)
	for i := 0; i < size; i++ {
		sem <- 1
	}
	return sem
}
