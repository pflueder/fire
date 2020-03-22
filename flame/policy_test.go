package flame

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/256dpi/fire/coal"
)

func TestPolicyIssueAndVerify(t *testing.T) {
	p := DefaultPolicy(testNotary)
	p.TokenData = func(c Client, ro ResourceOwner, t GenericToken) map[string]interface{} {
		return map[string]interface{}{
			"name": ro.(*User).Name,
		}
	}

	token := &Token{
		Base:      coal.B(),
		ExpiresAt: time.Now().Add(time.Hour).Round(time.Second),
	}

	sig, err := p.Issue(token, nil, &User{Name: "Hello"})
	assert.NoError(t, err)

	key, err := p.Verify(sig)
	assert.NoError(t, err)
	assert.Equal(t, token.ID(), key.ID)
	assert.Equal(t, token.ExpiresAt, key.Expiry.Local())
	assert.Equal(t, coal.Map{
		"name": "Hello",
	}, key.Extra)
}
