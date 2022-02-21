package tubes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProcess(t *testing.T) {
	assert := assert.New(t)

	var p *Process
	var err error
	p, err = NewProcess("go", "env")
	assert.NotNil(p)
	assert.NoError(err)

	d, err := p.RecvAll()
	assert.NotEmpty(d)
	assert.NoError(err)

	err = p.CloseFunc()
	assert.NoError(err)

	p, err = NewProcess("does", "not", "exist")
	assert.Nil(p)
	assert.Error(err)
}
