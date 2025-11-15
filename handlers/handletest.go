// Package handlers contains information about the handlers
package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/starfederation/datastar-go/datastar"
)

func HandleTest(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)
	updateCount := rand.Intn(10) + 3
	sse.PatchElements(fmt.Sprintf("<pre><code>Sending down %d numbers</code></pre>", updateCount),
		datastar.WithSelectorID("appendhere"), datastar.WithModeInner(), datastar.WithModeAppend())
	for range updateCount {
		sse.PatchElements(fmt.Sprintf("<pre><code>%d</code></pre>", rand.Intn(100)),
			datastar.WithSelectorID("appendhere"), datastar.WithModeInner(), datastar.WithModeAppend())
		time.Sleep(1000 * time.Millisecond)
	}
}
