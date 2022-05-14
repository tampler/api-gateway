package apiserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testBucket = "ec2buck"
	testPet    = "Teddy"
)

func Test_kv(t *testing.T) {

	// Connect to NATS
	nc, err := MakeNatsConnect()
	assert.NoError(t, err)

	js, err := nc.JetStream()
	assert.NoError(t, err)

	kv, err := js.KeyValue(testBucket)
	assert.NoError(t, err)

	kv.Put("pet", []byte(testPet))

	pet, err := kv.Get("pet")

	assert.Equal(t, testPet, string(pet.Value()))
}
