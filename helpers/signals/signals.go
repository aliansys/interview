// пакет для работы с сигналами (таймерами, внешними событиями)
package signals

import (
	"context"
	"os"
	"os/signal"
)

// перекрыть сигналы операционной системы и при их получении вернуть пустую структуру в порожденный канал
// usage:
// done := Signal(syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)
// <- done
func Signal(sigs ...os.Signal) <-chan struct{} {
	ctx, cancel := context.WithCancel(context.Background())
	var sigChan = make(chan os.Signal, 1)

	signal.Notify(sigChan, sigs...)

	go func() {
		<-sigChan
		cancel()
	}()
	return ctx.Done()

}
