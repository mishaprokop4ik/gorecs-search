package lexer_test

import (
	"github.com/mishaprokop4ik/gorecs-search/lexer"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewLexer(t *testing.T) {
	content :=
		`   Creating cgroups and moving processes
        A cgroup filesystem initially contains a single root cgroup,
        '/', which all processes belong to.  A new cgroup is created
        by creating a directory in the cgroup filesystem:

            mkdir /sys/fs/cgroup/cpu/cg1

        This creates a new empty cgroup.
        A process  may be moved  to this  cgroup by writing  its PID
        into the cgroup's cgroup.procs file:

            echo $$ > /sys/fs/cgroup/cpu/cg1/cgroup.procs

        Only one PID at a time should be written to this file.

`

	l := lexer.NewLexer(content)

	t.Log("Comparing input terms and in Lexer")
	assert.Equal(t, content, string(l.Terms))
}
