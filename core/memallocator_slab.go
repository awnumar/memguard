package core

import (
	"errors"
	"sort"
	"sync"
	"unsafe"
)

var (
	ErrSlabNoCacheFound = errors.New("no slab cache matching request")
	ErrSlabTooLarge     = errors.New("requested size too large")
)

type SlabAllocatorConfig struct {
	MinCanarySize int
	Sizes         []int
}

// Configuration options
type SlabOption func(*SlabAllocatorConfig)

// WithSizes allows to overwrite the SLAB Page sizes, defaulting to
// 64, 128, 256, 512, 1024 and 2048 byte
func WithSizes(sizes []int) SlabOption {
	return func(cfg *SlabAllocatorConfig) {
		cfg.Sizes = sizes
	}
}

// WithMinCanarySize allows to specify the minimum canary size (default: 16 byte)
func WithMinCanarySize(size int) SlabOption {
	return func(cfg *SlabAllocatorConfig) {
		cfg.MinCanarySize = size
	}
}

// Memory allocator implementation
type slabAllocator struct {
	maxSlabSize int
	cfg         *SlabAllocatorConfig
	allocator   *pageAllocator
	slabs       []*slab
}

func NewSlabAllocator(options ...SlabOption) MemAllocator {
	cfg := &SlabAllocatorConfig{
		MinCanarySize: 16,
		Sizes:         []int{64, 128, 256, 512, 1024, 2048},
	}
	for _, o := range options {
		o(cfg)
	}
	sort.Ints(cfg.Sizes)

	if len(cfg.Sizes) == 0 {
		return nil
	}

	// Setup the allocator and initialize the slabs
	a := &slabAllocator{
		maxSlabSize: cfg.Sizes[len(cfg.Sizes)-1],
		cfg:         cfg,
		slabs:       make([]*slab, 0, len(cfg.Sizes)),
		allocator: &pageAllocator{
			objects: make(map[int]*pageObject),
		},
	}
	for _, size := range cfg.Sizes {
		s := &slab{
			objSize:   size,
			allocator: a.allocator,
		}
		a.slabs = append(a.slabs, s)
	}

	return a
}

func (a *slabAllocator) Alloc(size int) ([]byte, error) {
	if size < 1 {
		return nil, ErrNullAlloc
	}

	// If the requested size is bigger than the largest slab, just malloc
	// the memory.
	requiredSlabSize := size + a.cfg.MinCanarySize
	if requiredSlabSize > a.maxSlabSize {
		return a.allocator.Alloc(size)
	}

	// Determine which slab to use depending on the size
	var s *slab
	for _, current := range a.slabs {
		if requiredSlabSize <= current.objSize {
			s = current
			break
		}
	}
	if s == nil {
		return nil, ErrSlabNoCacheFound
	}
	buf, err := s.alloc(size)
	if err != nil {
		return nil, err
	}

	// Trunc the buffer to the required size if requested
	return buf, nil
}

func (a *slabAllocator) Protect(buf []byte, readonly bool) error {
	// For the slab allocator, the data-slice is not identical to a memory page.
	// However, protection rules can only be applied to whole memory pages,
	// therefore protection of the data-slice is not supported by the slab
	// allocator.
	return nil
}

func (a *slabAllocator) Inner(buf []byte) []byte {
	if len(buf) == 0 {
		return nil
	}

	// If the buffer size is bigger than the largest slab, just free
	// the memory.
	size := len(buf) + a.cfg.MinCanarySize
	if size > a.maxSlabSize {
		return a.allocator.Inner(buf)
	}

	// Determine which slab to use depending on the size
	var s *slab
	for _, current := range a.slabs {
		if size <= current.objSize {
			s = current
			break
		}
	}
	if s == nil {
		Panic(ErrSlabNoCacheFound)
	}

	for _, c := range s.pages {
		if offset, contained := contains(c.buffer, buf); contained {
			return c.buffer[offset : offset+s.objSize]
		}
	}
	return nil
}

func (a *slabAllocator) Free(buf []byte) error {
	size := len(buf) + a.cfg.MinCanarySize

	// If the buffer size is bigger than the largest slab, just free
	// the memory.
	if size > a.maxSlabSize {
		return a.allocator.Free(buf)
	}

	// Determine which slab to use depending on the size
	var s *slab
	for _, current := range a.slabs {
		if size <= current.objSize {
			s = current
			break
		}
	}
	if s == nil {
		return ErrSlabNoCacheFound
	}

	return s.free(buf)
}

// *** INTERNAL FUNCTIONS *** //

// Page implementation
type slabObject struct {
	offset int
	next   *slabObject
}

type slabPage struct {
	used   int
	head   *slabObject
	canary []byte
	buffer []byte
}

func newPage(page []byte, size int) *slabPage {
	if size > len(page) || size < 1 {
		Panic(ErrSlabTooLarge)
	}

	// Determine the number of objects fitting into the page
	count := len(page) / size

	// Init the Page meta-data
	c := &slabPage{
		head:   &slabObject{},
		canary: page[len(page)-size:],
		buffer: page,
	}

	// Use the last object to create a canary prototype
	if err := Scramble(c.canary); err != nil {
		Panic(err)
	}

	// Initialize the objects
	last := c.head
	offset := size
	for i := 1; i < count-1; i++ {
		obj := &slabObject{offset: offset}
		last.next = obj
		offset += size
		last = obj
	}

	return c
}

// Slab is a container for all Pages serving the same size
type slab struct {
	objSize   int
	allocator *pageAllocator
	pages     []*slabPage
	sync.Mutex
}

func (s *slab) alloc(size int) ([]byte, error) {
	s.Lock()
	defer s.Unlock()

	// Find the fullest Page that isn't completely filled
	var c *slabPage
	for _, current := range s.pages {
		if current.head != nil && (c == nil || current.used > c.used) {
			c = current
		}
	}

	// No Page available, create a new one
	if c == nil {
		// Use the page allocator to get a new guarded memory page
		page, err := s.allocator.Alloc(pageSize - s.objSize)
		if err != nil {
			return nil, err
		}
		c = newPage(page, s.objSize)
		s.pages = append(s.pages, c)
	}

	// Remove the object from the free-list and increase the usage count
	obj := c.head
	c.head = c.head.next
	c.used++

	data := getBufferPart(c.buffer, obj.offset, size)
	canary := getBufferPart(c.buffer, obj.offset+size, s.objSize-size)

	// Fill in the remaining bytes with canary
	Copy(canary, c.canary)

	return data, nil
}

func contains(buf, obj []byte) (int, bool) {
	bb := uintptr(unsafe.Pointer(&buf[0]))
	be := uintptr(unsafe.Pointer(&buf[len(buf)-1]))
	o := uintptr(unsafe.Pointer(&obj[0]))

	if bb <= be {
		return int(o - bb), bb <= o && o < be
	}
	return int(o - be), be <= o && o < bb
}

func (s *slab) free(buf []byte) error {
	s.Lock()
	defer s.Unlock()

	// Find the Page containing the object
	var c *slabPage
	var cidx, offset int
	for i, current := range s.pages {
		diff, contained := contains(current.buffer, buf)
		if contained {
			c = current
			cidx = i
			offset = diff
			break
		}
	}
	if c == nil {
		return ErrBufferNotOwnedByAllocator
	}

	// Wipe the buffer including the canary check
	if err := s.wipe(c, offset, len(buf)); err != nil {
		return err
	}
	obj := &slabObject{
		offset: offset,
		next:   c.head,
	}
	c.head = obj
	c.used--

	// In case the Page is completely empty, we should remove it and
	// free the underlying memory
	if c.used == 0 {
		err := s.allocator.Free(c.buffer)
		if err != nil {
			return err
		}

		s.pages = append(s.pages[:cidx], s.pages[cidx+1:]...)
	}

	return nil
}

func (s *slab) wipe(page *slabPage, offset, size int) error {
	canary := getBufferPart(page.buffer, -s.objSize, s.objSize)
	inner := getBufferPart(page.buffer, offset, s.objSize)
	data := getBufferPart(page.buffer, offset, size)

	// Wipe data field
	Wipe(data)

	// Verify the canary
	if !Equal(inner[len(data):], canary[:size]) {
		return ErrBufferOverflow
	}

	// Wipe the memory
	Wipe(inner)

	return nil
}
