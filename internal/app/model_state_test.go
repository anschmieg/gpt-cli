package app

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestModelStateGetter(t *testing.T) {
    m := NewModel()
    assert.Equal(t, StateInput, m.State())

    // Transition via Update with ResponseMsg
    _, _ = m.Update(ResponseMsg{Content: "ok"})
    assert.Equal(t, StateResponse, m.State())
}

