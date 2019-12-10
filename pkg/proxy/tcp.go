package proxy

import (
	"context"
	"io"
	"net"
	"strings"
	"sync"

	"go.uber.org/zap"
)

func NewTCPServer(src, dest string) (*TCPServer, error) {
	srcAddr, err := net.ResolveTCPAddr("tcp", src)
	if err != nil {
		return nil, err
	}

	destAddr, err := net.ResolveTCPAddr("tcp", dest)
	if err != nil {
		return nil, err
	}

	return &TCPServer{
		src:  srcAddr,
		dest: destAddr,
		log:  zap.L().Named("proxy").Named("tcp"),
	}, nil
}

type TCPServer struct {
	src, dest *net.TCPAddr
	log       *zap.Logger
}

func (s *TCPServer) Serve(ctx context.Context) error {
	lis, err := net.ListenTCP("tcp", s.src)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		lis.Close()
	}()

	for {
		srcConn, err := lis.AcceptTCP()
		if err != nil {
			if ne, ok := err.(net.Error); ok {
				if ne.Temporary() {
					s.log.Warn("failed to accept tcp connection", zap.Error(err))
					continue
				}
			}
			if isErrorClosedConn(err) {
				select {
				case <-ctx.Done():
					return nil // already shutdowned
				default:
					// no-op
				}
			}
			s.log.Warn("failed to accept tcp connection", zap.Error(err))
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			s.handleConn(ctx, srcConn)
		}()
	}
}

func (s *TCPServer) handleConn(ctx context.Context, srcConn *net.TCPConn) {
	s.log.Debug("start handling the connection")
	defer s.log.Debug("finish handling the connection")

	cp := func(src, dest *net.TCPConn, wg *sync.WaitGroup) {
		defer wg.Done()
		_, err := io.Copy(dest, src)
		if err != nil && !isErrorClosedConn(err) {
			// TODO: handle errors
			s.log.Warn("failed to copy packets", zap.Error(err))
			return
		}
	}

	destConn, err := net.DialTCP("tcp", nil, s.dest)
	if err != nil {
		// TODO: handle error
		return
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		srcConn.Close()
		destConn.Close()
	}()

	wg.Add(2)
	go cp(srcConn, destConn, &wg)
	go cp(destConn, srcConn, &wg)
}

func isErrorClosedConn(err error) bool {
	return strings.Contains(err.Error(), "use of closed network connection")
}
