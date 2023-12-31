package cpu

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStat(t *testing.T) {
	time.Sleep(time.Second * 2)
	var s Stat
	var i Info
	ReadStat(&s)
	i = GetInfo()
	
	fmt.Println(s, i)

	assert.NotZero(t, s.Usage)
	//assert.NotZero(t, i.Frequency)
	//assert.NotZero(t, i.Quota)
}
