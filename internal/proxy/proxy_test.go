package proxy

import (
	"fmt"
	"testing"
)

func TestProxyJwt(t *testing.T) {
	jwt, _ := proxyMock.Jwt()
	var z string
	t.Run("must be empty when no header", func(t *testing.T) {
		if jwt != z {
			t.Errorf("jwt must be empty, got %s", jwt)
		}
	})

	t.Run("must extract from header", func(t *testing.T) {
		expected := "jwt_mocked"
		proxyMock.Ctx.Req().Header.Set("Authorization", fmt.Sprintf("Bearer %s", expected))
		jwt, _ = proxyMock.Jwt()
		if jwt != expected {
			t.Errorf("jwt must be equal to %s, got %s", expected, jwt)
		}
	})
}

func TestProxyMustSplitMiddleware(t *testing.T) {
	before, after := proxyMock.splitMiddlewares()
	t.Run("before middleware must contain configured one", func(t *testing.T) {
		expected := 1
		if len(before) != 1 {
			t.Errorf("before must contain %d middleware configured got %d", expected, len(before))
		}
	})

	t.Run("after middleware must contain configured one", func(t *testing.T) {
		expected := 1
		if len(after) != 1 {
			t.Errorf("after must contain %d middleware configured got %d", expected, len(after))
		}
	})
}
