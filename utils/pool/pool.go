package pool

// Pool is a byte pool to avoid scrappy allocation
type Pool struct {
	pos int
	buf []byte
}

const maxpoolsize = 500 * 1024

// Get gets a byte slice of specific size
func (pool *Pool) Get(size int) []byte {
	if maxpoolsize-pool.pos < size {
		pool.pos = 0
		pool.buf = make([]byte, maxpoolsize)
	}
	b := pool.buf[pool.pos : pool.pos+size]
	pool.pos += size
	return b
}

// NewPool return a Pool
func NewPool() *Pool {
	return &Pool{
		buf: make([]byte, maxpoolsize),
	}
}
