package memguard

// The subclave container is similar to a normal container but it is only used
// internally to protect values that are used in the protection of normal containers.
type subclave struct {
	x []byte
	y []byte
}

func newSubclave() {

}
