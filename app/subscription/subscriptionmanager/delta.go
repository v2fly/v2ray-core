package subscriptionmanager

type changedDocument struct {
	removed   []string
	added     []string
	modified  []string
	unchanged []string
}
